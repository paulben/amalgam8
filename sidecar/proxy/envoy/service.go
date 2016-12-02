package envoy

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/Sirupsen/logrus"
)

// Service manages proxy service
type Service interface {
	Start() error
	Reload() error
	Stop() error
}

type service struct {
	config string
	cmdMap map[*exec.Cmd]int
	mutex  sync.Mutex
}

// NewService creates new instance
func NewService(config string) Service {
	return &service{
		config: config,
		cmdMap: make(map[*exec.Cmd]int),
	}
}

// Start
func (s *service) Start() error {
	//if s.cmd != nil {
	//	return nil
	//}
	//
	//s.cmd = exec.Command("envoy", "-c", s.config, "--restart-epoch", fmt.Sprint(s.restartEpoch))
	//
	//return s.cmd.Start()
	return nil
}

func (s *service) waitForExit(cmd *exec.Cmd) {
	if err := cmd.Wait(); err != nil {
		logrus.Infof("Envoy terminated: %v", err.Error())
	} else {
		logrus.Info("Envoy process exited with status 0")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.cmdMap, cmd)
}

// Reload
func (s *service) Reload() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	restartEpoch := -1
	for _, epoch := range s.cmdMap {
		if epoch > restartEpoch {
			restartEpoch = epoch
		}
	}
	restartEpoch++

	cmd := exec.Command("envoy", "-c", s.config, "--restart-epoch", fmt.Sprint(restartEpoch))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return err
	}
	s.cmdMap[cmd] = restartEpoch

	go s.waitForExit(cmd)

	return nil
}

// Stop Envoy
// Envoy propagates the SIGTERM to all running instances of Envoy (I.E., Envoy instances with lower
// restart-epoch numbers.)
func (s *service) Stop() error {
	//if err := s.cmd.Process.Signal(syscall.SIGTERM); err != nil {
	//	return err
	//}
	//return s.cmd.Process.Release()
	return nil
}
