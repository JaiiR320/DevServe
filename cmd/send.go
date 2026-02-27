package cmd

import (
	"devserve/daemon"
	"devserve/internal"
	"errors"
	"fmt"
)

// sendRequest sends a request to the daemon, auto-starting it if needed.
func sendRequest(req *internal.Request) (*internal.Response, error) {
	resp, err := internal.Send(req)
	if err == nil {
		return resp, nil
	}

	// Only auto-start if the daemon isn't running
	if !errors.Is(err, internal.ErrDaemonNotRunning) {
		return resp, err
	}

	// Auto-start the daemon
	if startErr := daemon.Start(true); startErr != nil {
		return nil, fmt.Errorf("failed to auto-start daemon: %w", startErr)
	}

	// Retry the request
	return internal.Send(req)
}
