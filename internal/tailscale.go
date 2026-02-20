package internal

import (
	"io"
	"os/exec"
	"strconv"
)

type TailscaleManager struct {
	Stdout io.Writer
	Stderr io.Writer
}

func NewTailscaleManager(stdout, stderr io.Writer) *TailscaleManager {
	return &TailscaleManager{
		Stdout: stdout,
		Stderr: stderr,
	}
}

func (tm *TailscaleManager) Start(port int) error {
	portStr := strconv.Itoa(port)

	cmd := exec.Command("tailscale", "serve", "--https", portStr, "--bg", "--yes", portStr)

	cmd.Stdout = tm.Stdout
	cmd.Stderr = tm.Stderr

	return cmd.Run()
}

func (tm *TailscaleManager) Stop(port int) error {
	portStr := strconv.Itoa(port)
	cmd := exec.Command("tailscale", "serve", "--https", portStr, "off")

	cmd.Stdout = tm.Stdout
	cmd.Stderr = tm.Stderr

	return cmd.Run()
}
