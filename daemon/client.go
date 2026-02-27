package daemon

import (
	"devserve/config"
	"errors"
	"fmt"
	"net"
)

// ErrDaemonNotRunning is returned when the daemon socket cannot be reached.
var ErrDaemonNotRunning = errors.New("daemon is not running")

func Send(req *Request) (*Response, error) {
	conn, err := net.Dial("unix", config.Socket)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDaemonNotRunning, err)
	}
	defer conn.Close()

	err = SendRequest(conn, req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return ReadResponse(conn)
}
