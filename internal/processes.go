package internal

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

type Process struct {
	Cmd            *exec.Cmd `json:"-"`
	Port           int       `json:"port"`
	PackageManager string    `json:"package_manager"`
	PID            int       `json:"pid"`
}

func CreateProcess(port int, pm string, command string, args ...string) *Process {
	return &Process{
		Cmd:            exec.Command(command, args...),
		Port:           port,
		PackageManager: pm,
	}
}

func (p *Process) SetOutputs(stdout, stderr io.Writer) {
	p.Cmd.Stdout = stdout
	p.Cmd.Stderr = stderr
}

func (p *Process) Start() error {
	p.Cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	sigChan := make(chan os.Signal, 1)
	doneChan := make(chan struct{}, 1)

	signal.Ignore(os.Interrupt)
	signal.Notify(sigChan, os.Interrupt)
	err := p.Cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start dev server: %w", err)
	}

	p.PID = p.Cmd.Process.Pid

	go func() {
		p.Cmd.Wait()
		doneChan <- struct{}{}
	}()

	select {
	case <-sigChan:
		p.Stop()
	case <-doneChan:
	}
	return nil
}

func (p *Process) StartBG(fm *FileManager) error {
	err := p.Cmd.Start()
	if err != nil {
		return fmt.Errorf("Failed starting command: %w", err)
	}
	p.PID = p.Cmd.Process.Pid

	saveProcess(fm, p)

	return nil
}

func (p *Process) Stop() {
	syscall.Kill(-p.PID, syscall.SIGTERM)
	time.Sleep(100 * time.Millisecond)
	syscall.Kill(-p.PID, syscall.SIGKILL)
}
