package internal

import (
	"devserve/util"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type Process struct {
	Name   string
	Cmd    *exec.Cmd
	Port   int
	Dir    string
	Stdout *os.File
	Stderr *os.File

	mu            sync.Mutex
	started       bool
	stopped       bool
	processKilled bool
}

// initialize out files, and process struct
func CreateProcess(name string, port int, dir string) (*Process, error) {
	logDir := filepath.Join(dir, util.ProcessLogDir)
	if err := os.MkdirAll(logDir, util.DirPermissions); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	outFile, err := os.Create(filepath.Join(logDir, util.ProcessStdoutLog))
	if err != nil {
		return nil, fmt.Errorf("failed to create output log: %w", err)
	}
	errFile, err := os.Create(filepath.Join(logDir, util.ProcessStderrLog))
	if err != nil {
		return nil, fmt.Errorf("failed to create error log: %w", err)
	}

	return &Process{
		Name:   name,
		Port:   port,
		Dir:    dir,
		Stdout: outFile,
		Stderr: errFile,
	}, nil
}

func (p *Process) Start(command string) error {
	p.Cmd = exec.Command("sh", "-c", command)

	p.Cmd.Stderr = p.Stderr
	p.Cmd.Stdout = p.Stdout
	if p.Dir != "" {
		p.Cmd.Dir = p.Dir
	}

	p.Cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	err := p.Cmd.Start()
	if err != nil {
		p.closeLogs()
		return fmt.Errorf("failed to start command: %w", err)
	}

	p.mu.Lock()
	p.started = true
	p.mu.Unlock()

	log.Printf("waiting for port %d...", p.Port)
	if err := WaitForPort(p.Port, util.PortWaitTimeout); err != nil {
		syscall.Kill(-p.Cmd.Process.Pid, syscall.SIGTERM)
		p.closeLogs()
		return fmt.Errorf("failed to wait for port %d: %w", p.Port, err)
	}

	if err := DefaultTunnel.Serve(p.Port); err != nil {
		sysErr := syscall.Kill(-p.Cmd.Process.Pid, syscall.SIGTERM)
		p.closeLogs()
		if sysErr != nil {
			return fmt.Errorf("failed to kill process after tailscale error: %w", sysErr)
		}
		return fmt.Errorf("failed to enable tailscale serve: %w", err)
	}
	return nil
}

func (p *Process) Stop() error {
	p.mu.Lock()
	if !p.started {
		p.mu.Unlock()
		return fmt.Errorf("process '%s' has not been started", p.Name)
	}
	if p.stopped {
		p.mu.Unlock()
		return fmt.Errorf("process '%s' is already stopped", p.Name)
	}
	needsKill := !p.processKilled
	p.mu.Unlock()

	if needsKill {
		log.Printf("stopping process %s (pid %d)", p.Name, p.Cmd.Process.Pid)
		err := syscall.Kill(-p.Cmd.Process.Pid, syscall.SIGTERM)
		if err != nil {
			return fmt.Errorf("failed to send SIGTERM to process '%s': %w", p.Name, err)
		}

		// Wait for the process to exit with a 5s timeout, escalate to SIGKILL if needed
		done := make(chan error, 1)
		go func() {
			done <- p.Cmd.Wait()
		}()

		select {
		case <-done:
			log.Printf("process %s exited gracefully", p.Name)
		case <-time.After(util.StopGracePeriod):
			log.Printf("process %s did not exit after SIGTERM, sending SIGKILL", p.Name)
			if killErr := syscall.Kill(-p.Cmd.Process.Pid, syscall.SIGKILL); killErr != nil {
				log.Printf("failed to SIGKILL process %s: %s", p.Name, killErr)
			}
			<-done // wait for Wait() to return after SIGKILL
			log.Printf("process %s killed with SIGKILL", p.Name)
		}

		p.closeLogs()

		p.mu.Lock()
		p.processKilled = true
		p.mu.Unlock()
	}

	if err := DefaultTunnel.Stop(p.Port); err != nil {
		return fmt.Errorf("failed to disable tailscale serve: %w", err)
	}

	p.mu.Lock()
	p.stopped = true
	p.mu.Unlock()
	return nil
}

func (p *Process) closeLogs() {
	if p.Stdout != nil {
		p.Stdout.Close()
	}
	if p.Stderr != nil {
		p.Stderr.Close()
	}
}
