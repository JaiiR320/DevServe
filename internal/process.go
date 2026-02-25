package internal

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
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
	log.Println("starting process", p.Name)
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
	log.Println("starting tailscale on port", portStr)
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
	log.Println("stopping process", p.Name)
	err := syscall.Kill(-p.Cmd.Process.Pid, syscall.SIGTERM)
	if err != nil {
		return err
	}

	p.Stdout.Close()
	p.Stderr.Close()

	portStr := strconv.Itoa(p.Port)
	log.Println("stopping tailscale")
	cmd := exec.Command("tailscale", "serve", "--https", portStr, "off")
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
