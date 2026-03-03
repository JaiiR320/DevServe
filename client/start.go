package client

import (
	"devserve/config"
	"devserve/protocol"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

// StartDaemon starts the daemon in the background.
// It creates the log directory, re-executes the binary with "daemon start --foreground",
// waits briefly, then verifies the daemon is running with a ping.
// Returns an error if the daemon is already running or fails to start.
func StartDaemon() error {
	err := os.MkdirAll(config.DaemonDir, config.DirPermissions)
	if err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logFile, err := os.Create(filepath.Join(config.DaemonDir, config.DaemonLogFile))
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer logFile.Close()

	conn, err := net.Dial("unix", config.Socket)
	if err == nil {
		conn.Close()
		return errors.New("daemon is already running")
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	cmd := exec.Command(execPath, "daemon", "start", "--foreground")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	// Wait a moment and verify daemon started with ping
	time.Sleep(config.DaemonStartDelay)
	pingReq := &protocol.Request{Action: "ping"}
	resp, err := Send(pingReq)
	if err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}
	if !resp.OK || resp.Data != "pong" {
		return errors.New("daemon health check failed")
	}

	return nil
}
