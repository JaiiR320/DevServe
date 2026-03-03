package client

import (
	"devserve/config"
	"devserve/protocol"
	"errors"
	"fmt"
	"net"
)

// ErrDaemonNotRunning is returned when the daemon socket cannot be reached.
var ErrDaemonNotRunning = errors.New("daemon is not running")

// Send sends a request to the daemon and returns the response.
func Send(req *protocol.Request) (*protocol.Response, error) {
	conn, err := net.Dial("unix", config.Socket)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDaemonNotRunning, err)
	}
	defer conn.Close()

	err = protocol.SendRequest(conn, req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return protocol.ReadResponse(conn)
}
