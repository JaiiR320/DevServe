package internal

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
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

func (p *Process) Start(fm *FileManager) error {
	p.Cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	fmt.Println("Starting dev server...")
	err := p.Cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start dev server: %w", err)
	}

	p.PID = p.Cmd.Process.Pid

	portstr := strconv.Itoa(p.Port)

	attempts := 0
	for {
		if attempts == 60 {
			p.Stop()
			return fmt.Errorf("Max attempts to dial")
		}
		conn, _ := net.Dial("tcp", "localhost:"+portstr)
		if conn != nil {
			conn.Close()
			break
		}
		time.Sleep(250 * time.Millisecond)
		attempts++
	}

	err = saveProcess(fm, p)
	if err != nil {
		p.Stop()
		return fmt.Errorf("Failed to save process: %w", err)
	}
	fmt.Println("Dev server started. Listening on " + portstr)
	return nil
}

func (p *Process) Wait() error {
	sigChan := make(chan os.Signal, 1)
	doneChan := make(chan struct{}, 1)

	signal.Ignore(os.Interrupt)
	signal.Notify(sigChan, os.Interrupt)

	go func() {
		p.Cmd.Wait()
		doneChan <- struct{}{}
	}()

	// wait for one of these conditions
	select {
	case <-sigChan:
	case <-doneChan:
	}

	p.Stop()
	return nil
}

func (p *Process) Stop() {
	syscall.Kill(-p.PID, syscall.SIGTERM)
	time.Sleep(100 * time.Millisecond)
	syscall.Kill(-p.PID, syscall.SIGKILL)
}
