package internal

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func Serve(port int, bg bool) error {
	portStr := strconv.Itoa(port)

	pm, err := DetectPackageManager()
	if err != nil {
		return err
	}

	var stdout io.Writer
	var stderr io.Writer

	fm := NewLocalFM()
	if bg {
		err = fm.InitDir()
		if err != nil {
			return fmt.Errorf("Failed to init directory: %w", err)
		}
		stdout, stderr, err = fm.CreateLogFiles()
		if err != nil {
			return fmt.Errorf("Failed creating std log files: %w", err)
		}
	} else {
		stdout = os.Stdout
		stderr = os.Stderr
	}

	tm := NewTailscaleManager(stdout, stderr)

	err = tm.Start(portStr)
	if err != nil {
		return fmt.Errorf("failed to start Tailscale: %w", err)
	}

	command := PMToCommand[pm]
	args := strings.Split(command, " ")
	cmd := exec.Command(args[0], args[1:]...)

	cmd.Stderr = stderr
	cmd.Stdout = stdout

	if bg {
		err = cmd.Start()
		if err != nil {
			return err
		}
		pid := cmd.Process.Pid
		err = fm.SavePID(pid)
		if err != nil {
			return fmt.Errorf("failed saving PID: %w", err)
		}

		return nil
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	sigChan := make(chan os.Signal, 1)
	doneChan := make(chan struct{}, 1)

	signal.Ignore(os.Interrupt)
	signal.Notify(sigChan, os.Interrupt)
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start dev server: %w", err)
	}

	go func() {
		cmd.Wait()
		doneChan <- struct{}{}
	}()

	select {
	case <-sigChan:
		// Kill entire process group
		syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
		time.Sleep(100 * time.Millisecond)
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		return tm.Stop(portStr)
	case <-doneChan:
		return tm.Stop(portStr)
	}
}
