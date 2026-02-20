package internal

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

func Serve(port int) error {
	pm, err := DetectPackageManager()
	if err != nil {
		return err
	}
	portStr := strconv.Itoa(port)
	cmd := exec.Command("tailscale", "serve", "--https", portStr, "--bg", "--yes", portStr)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("tailscale failed: %w", err)
	}

	command := PMToCommand[pm]
	args := strings.Split(command, " ")
	cmd = exec.Command(args[0], args[1:]...)

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	sigChan := make(chan os.Signal, 1)
	doneChan := make(chan struct{}, 1)

	fmt.Println("\nstarting dev server")
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
		err := cmd.Process.Kill()
		if err != nil {
			return err
		}
		return stopTailscale(portStr)
	case <-doneChan:
		fmt.Println("dev server exited")
		return stopTailscale(portStr)
	}
}

func stopTailscale(portStr string) error {
	cmd := exec.Command("tailscale", "serve", "--https", portStr, "off")
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
