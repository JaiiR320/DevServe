package internal

import (
	"io"
	"os/exec"
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

func (tm *TailscaleManager) Start(port string) error {
	cmd := exec.Command("tailscale", "serve", "--https", port, "--bg", "--yes", port)

	cmd.Stdout = tm.Stdout
	cmd.Stderr = tm.Stderr

	return cmd.Run()
}

func (tm *TailscaleManager) Stop(port string) error {
	cmd := exec.Command("tailscale", "serve", "--https", port, "off")

	cmd.Stdout = tm.Stdout
	cmd.Stderr = tm.Stderr

	return cmd.Run()
}
