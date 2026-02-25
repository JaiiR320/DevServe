package internal

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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
}

// initialize out files, and process struct
func CreateProcess(name string, port int, dir string) (*Process, error) {
	logDir := filepath.Join(dir, ".devserve")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	outFile, err := os.Create(filepath.Join(logDir, "out.log"))
	if err != nil {
		return nil, err
	}
	errFile, err := os.Create(filepath.Join(logDir, "err.log"))
	if err != nil {
		return nil, err
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
	args := strings.Split(command, " ")
	p.Cmd = exec.Command(args[0], args[1:]...)

	p.Cmd.Stderr = p.Stderr
	p.Cmd.Stdout = p.Stdout
	if p.Dir != "" {
		p.Cmd.Dir = p.Dir
	}

	p.Cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	err := p.Cmd.Start()
	if err != nil {
		return err
	}

	log.Printf("waiting for port %d...", p.Port)
	if err := WaitForPort(p.Port, 15*time.Second); err != nil {
		syscall.Kill(-p.Cmd.Process.Pid, syscall.SIGTERM)
		return err
	}

	portStr := strconv.Itoa(p.Port)
	cmd := exec.Command("tailscale", "serve", "--https", portStr, "--bg", "localhost:"+portStr)
	err = cmd.Run()
	if err != nil {
		sysErr := syscall.Kill(-p.Cmd.Process.Pid, syscall.SIGTERM)
		if sysErr != nil {
			return sysErr
		}
		return err
	}
	return nil
}

func (p *Process) Stop() error {
	log.Printf("stopping process %s (pid %d)", p.Name, p.Cmd.Process.Pid)
	err := syscall.Kill(-p.Cmd.Process.Pid, syscall.SIGTERM)
	if err != nil {
		return err
	}

	// Wait for the process to exit with a 5s timeout, escalate to SIGKILL if needed
	done := make(chan error, 1)
	go func() {
		done <- p.Cmd.Wait()
	}()

	select {
	case <-done:
		log.Printf("process %s exited gracefully", p.Name)
	case <-time.After(5 * time.Second):
		log.Printf("process %s did not exit after SIGTERM, sending SIGKILL", p.Name)
		if killErr := syscall.Kill(-p.Cmd.Process.Pid, syscall.SIGKILL); killErr != nil {
			log.Printf("failed to SIGKILL process %s: %s", p.Name, killErr)
		}
		<-done // wait for Wait() to return after SIGKILL
		log.Printf("process %s killed with SIGKILL", p.Name)
	}

	p.Stdout.Close()
	p.Stderr.Close()

	portStr := strconv.Itoa(p.Port)
	cmd := exec.Command("tailscale", "serve", "--https", portStr, "off")
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
