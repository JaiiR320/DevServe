package internal

import (
	"fmt"
	"log"
	"net"
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

func CheckPortAvailable(port int) error {
	addr := "localhost:" + strconv.Itoa(port)
	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err == nil {
		conn.Close()
		return fmt.Errorf("port %d is already in use", port)
	}
	return nil
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
		// process exited
	case <-time.After(5 * time.Second):
		log.Printf("process %s did not exit after SIGTERM, sending SIGKILL", p.Name)
		if killErr := syscall.Kill(-p.Cmd.Process.Pid, syscall.SIGKILL); killErr != nil {
			log.Printf("failed to SIGKILL process %s: %s", p.Name, killErr)
		}
		<-done // wait for Wait() to return after SIGKILL
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
