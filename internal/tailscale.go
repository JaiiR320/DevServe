package internal

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// TailscaleInfo holds the hostname and IP address from Tailscale status.
type TailscaleInfo struct {
	Hostname string
	IP       string
}

// CommandRunner is a function that returns the output of a command.
type CommandRunner func() ([]byte, error)

// DefaultTailscaleRunner runs `tailscale status --json`.
func DefaultTailscaleRunner() ([]byte, error) {
	return exec.Command("tailscale", "status", "--json").Output()
}

// GetTailscaleInfo parses tailscale status JSON using the provided runner.
func GetTailscaleInfo(runner CommandRunner) (*TailscaleInfo, error) {
	data, err := runner()
	if err != nil {
		return nil, fmt.Errorf("failed to get tailscale status: %w", err)
	}

	var status struct {
		TailscaleIPs []string `json:"TailscaleIPs"`
		Self         struct {
			DNSName string `json:"DNSName"`
		} `json:"Self"`
	}

	if err := json.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("failed to parse tailscale status: %w", err)
	}

	info := &TailscaleInfo{}

	if len(status.TailscaleIPs) > 0 {
		info.IP = status.TailscaleIPs[0]
	}

	info.Hostname = strings.TrimSuffix(status.Self.DNSName, ".")

	return info, nil
}
