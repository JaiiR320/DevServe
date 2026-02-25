package internal

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

type Process struct {
	Name   string
	Cmd    *exec.Cmd
	Port   int
	Stdout *os.File
	Stderr *os.File
}

// initialize out files, and process struct
func CreateProcess(name string, port int) (*Process, error) {
	outFile, err := os.Create("/tmp/" + name + ".out.log")
	if err != nil {
		return nil, err
	}
	errFile, err := os.Create("/tmp/" + name + ".err.log")
	if err != nil {
		return nil, err
	}

	return &Process{
		Name:   name,
		Port:   port,
		Stdout: outFile,
		Stderr: errFile,
	}, nil
}

func (p *Process) Start(command string) error {
	fmt.Println("Starting process", p.Name)
	args := strings.Split(command, " ")
	p.Cmd = exec.Command(args[0], args[1:]...)

	p.Cmd.Stderr = p.Stderr
	p.Cmd.Stdout = p.Stdout

	p.Cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	err := p.Cmd.Start()
	if err != nil {
		return err
	}

	portStr := strconv.Itoa(p.Port)
	fmt.Println("Starting tailscale on port", portStr)
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
	fmt.Println("Stopping process", p.Name)
	err := syscall.Kill(-p.Cmd.Process.Pid, syscall.SIGTERM)
	if err != nil {
		return err
	}

	p.Stdout.Close()
	p.Stderr.Close()

	portStr := strconv.Itoa(p.Port)
	fmt.Println("Stopping tailscale")
	cmd := exec.Command("tailscale", "serve", "--https", portStr, "off")
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
