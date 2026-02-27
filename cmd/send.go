package cmd

import (
	"devserve/daemon"
	"errors"
	"fmt"
)

// sendRequest sends a request to the daemon, auto-starting it if needed.
func sendRequest(req *daemon.Request) (*daemon.Response, error) {
	resp, err := daemon.Send(req)
	if err == nil {
		return resp, nil
	}

	// Only auto-start if the daemon isn't running
	if !errors.Is(err, daemon.ErrDaemonNotRunning) {
		return resp, err
	}

	// Auto-start the daemon
	if startErr := daemon.Start(true); startErr != nil {
		return nil, fmt.Errorf("failed to auto-start daemon: %w", startErr)
	}

	// Retry the request
	return daemon.Send(req)
}
