package client

import (
	"fmt"

	"github.com/jaiir320/devserve/config"
	"github.com/jaiir320/devserve/process"
	"github.com/jaiir320/devserve/protocol"
)

// Restart stops a running process, waits for its port to be released, and
// starts it again using its existing live configuration.
func Restart(name string) (*protocol.ServeResult, error) {
	info, err := Get(name)
	if err != nil {
		return nil, fmt.Errorf("failed to query process '%s': %w", name, err)
	}

	if err := Stop(name); err != nil {
		return nil, fmt.Errorf("failed to stop: %w", err)
	}

	if err := process.WaitForPortFree(info.Port, config.PortWaitTimeout); err != nil {
		return nil, fmt.Errorf("failed waiting for port %d to be released: %w", info.Port, err)
	}

	result, err := Serve(info.Name, info.Port, info.Command, info.Dir)
	if err != nil {
		return nil, fmt.Errorf("failed to start: %w", err)
	}

	return result, nil
}
