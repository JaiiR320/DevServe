package internal

import (
	"fmt"
	"net"
)

// Send a request to the daemon and return the response
func Send(req *Request) (*Response, error) {
	conn, err := net.Dial("unix", Socket)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	err = SendRequest(conn, req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return ReadResponse(conn)
}
