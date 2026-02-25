package internal

import (
	"net"
)

// Send a request to the daemon and return the response
func Send(req *Request) (*Response, error) {
	conn, err := net.Dial("unix", Socket)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	err = SendRequest(conn, req)
	if err != nil {
		return nil, err
	}

	return ReadResponse(conn)
}
