package cmd

import (
	"devserve/client"
	"devserve/protocol"
	"errors"
	"fmt"
)

// sendRequest sends a request to the daemon, auto-starting it if needed.
func sendRequest(req *protocol.Request) (*protocol.Response, error) {
	resp, err := client.Send(req)
	if err == nil {
		return resp, nil
	}

	// Only auto-start if the daemon isn't running
	if !errors.Is(err, client.ErrDaemonNotRunning) {
		return resp, err
	}

	// Auto-start the daemon
	if startErr := client.StartDaemon(); startErr != nil {
		return nil, fmt.Errorf("failed to auto-start daemon: %w", startErr)
	}

	// Retry the request
	return client.Send(req)
}
