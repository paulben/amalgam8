package envoy

import (
	"fmt"
	"os/exec"
	"syscall"
)

// Service manages proxy service
type Service interface {
	Start() error
	Reload() error
	Stop() error
}

type service struct {
	cmd          *exec.Cmd
	config       string
	restartEpoch int
}

// NewService creates new instance
func NewService(config string) Service {
	return &service{
		config: config,
	}
}

// Start
func (s *service) Start() error {
	if s.cmd != nil {
		return nil
	}

	s.cmd = exec.Command("envoy", "-c", s.config, "--restart-epoch", fmt.Sprint(s.restartEpoch))
	return s.cmd.Start()
}

// Reload
func (s *service) Reload() error {
	if s.cmd != nil {
		s.cmd.Process.Release()
	}

	s.cmd = exec.Command("envoy", "-c", s.config, "--restart-epoch", fmt.Sprint(s.restartEpoch))
	s.restartEpoch++

	return s.cmd.Start()
}

// Stop Envoy
// Envoy propagates the SIGTERM to all running instances of Envoy (I.E., Envoy instances with lower
// restart-epoch numbers.)
func (s *service) Stop() error {
	if err := s.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return err
	}
	return s.cmd.Process.Release()
}
