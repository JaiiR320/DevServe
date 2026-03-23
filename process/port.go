package process

import (
	"fmt"
	"github.com/jaiir320/devserve/config"
	"net"
	"strconv"
	"time"
)

func CheckPortInUse(port int) error {
	addr := "localhost:" + strconv.Itoa(port)
	conn, err := net.DialTimeout("tcp", addr, config.PortDialTimeout)
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
			conn, err := net.DialTimeout("tcp", addr, config.PortDialTimeout)
			if err == nil {
				conn.Close()
				return nil
			}
			time.Sleep(config.PortPollInterval)
		}
	}
}

func WaitForPortFree(port int, timeout time.Duration) error {
	deadline := time.After(timeout)
	for {
		select {
		case <-deadline:
			return fmt.Errorf("port %d still in use after %s", port, timeout)
		default:
			if err := CheckPortInUse(port); err == nil {
				return nil
			}
			time.Sleep(config.PortPollInterval)
		}
	}
}
