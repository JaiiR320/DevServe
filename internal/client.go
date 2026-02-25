package internal

import (
	"net"
)

// Send a message to the daemon
func Send(msg string) error {
	conn, err := net.Dial("unix", Socket)
	if err != nil {
		return err
	}
	defer conn.Close()

	data := []byte(msg)

	_, err = conn.Write(data)
	if err != nil {
		return err
	}
	return nil
}
