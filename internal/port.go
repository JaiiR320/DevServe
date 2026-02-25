package internal

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

func CheckPortAvailable(port int) error {
	addr := "localhost:" + strconv.Itoa(port)
	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err == nil {
		conn.Close()
		return fmt.Errorf("port %d is already in use", port)
	}
	return nil
}

func WaitForPort(port int, timeout time.Duration) error {
	addr := "localhost:" + strconv.Itoa(port)
	deadline := time.After(timeout)
	for {
		select {
		case <-deadline:
			return fmt.Errorf("port %d not ready after %s", port, timeout)
		default:
			conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
			if err == nil {
				conn.Close()
				return nil
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
}
