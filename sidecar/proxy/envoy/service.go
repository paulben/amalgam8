package envoy

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/Sirupsen/logrus"
)

// Service manages a proxy service.
type Service interface {
	Reload() error
}

type service struct {
	config string            // Envoy config filename.
	cmdMap map[*exec.Cmd]int // Map of known running Envoy processes and their restart epochs.
	mutex  sync.Mutex
}

// NewService creates new instance.
func NewService(config string) Service {
	return &service{
		config: config,
		cmdMap: make(map[*exec.Cmd]int),
	}
}

// Reload Envoy with a hot restart. Envoy hot restarts are performed by launching a new Envoy process with an
// incremented restart epoch. To successfully launch a new Envoy process that will replace the running Envoy processes,
// the restart epoch of the new process must be exactly 1 greater than the highest restart epoch of the currently
// running Envoy processes. To ensure that we launch the new Envoy process with the correct restart epoch, we keep track
// of all running Envoy processes and their restart epochs.
//
// Envoy hot restart documentation: https://lyft.github.io/envoy/docs/intro/arch_overview/hot_restart.html
func (s *service) Reload() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Find highest restart epoch of the known running Envoy processes.
	restartEpoch := -1
	for _, epoch := range s.cmdMap {
		if epoch > restartEpoch {
			restartEpoch = epoch
		}
	}
	restartEpoch++

	// Spin up a new Envoy process.
	cmd := exec.Command("envoy", "-c", s.config, "--restart-epoch", fmt.Sprint(restartEpoch))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}

	// Add the new Envoy process to the known set of running Envoy processes.
	s.cmdMap[cmd] = restartEpoch

	// Start tracking the process.
	go s.waitForExit(cmd)

	return nil
}

// waitForExit waits until the command exits and removes it from the set of known running Envoy processes.
func (s *service) waitForExit(cmd *exec.Cmd) {
	if err := cmd.Wait(); err != nil {
		logrus.Debugf("Envoy terminated: %v", err.Error())
	} else {
		logrus.Debug("Envoy process exited")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.cmdMap, cmd)
}
