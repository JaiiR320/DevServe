package internal

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
)

func Serve(port int, bg bool) error {
	portStr := strconv.Itoa(port)

	pm, err := DetectPackageManager()
	if err != nil {
		return err
	}

	stdout, stderr, err := getOutputs(bg)

	cmd := exec.Command("tailscale", "serve", "--https", portStr, "--bg", "--yes", portStr)

	cmd.Stderr = stderr
	cmd.Stdout = stdout

	err = cmd.Run()
	if err != nil {
		return err
	}

	command := PMToCommand[pm]
	args := strings.Split(command, " ")
	cmd = exec.Command(args[0], args[1:]...)

	cmd.Stderr = stderr
	cmd.Stdout = stdout

	if bg {
		return cmd.Start()
	}

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

func getOutputs(bg bool) (outWriter io.Writer, errWriter io.Writer, error error) {
	if !bg {
		return os.Stdout, os.Stderr, nil
	}

	err := os.MkdirAll(".devserve", 0755)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create folder")
	}
	outFile, err := os.Create(".devserve/out.log")
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't create out.log")
	}

	errFile, err := os.Create(".devserve/err.log")
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't create out.log")
	}

	return outFile, errFile, nil
}

func stopTailscale(portStr string) error {
	cmd := exec.Command("tailscale", "serve", "--https", portStr, "off")
	return cmd.Run()
}
