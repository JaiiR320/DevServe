package client

import (
	"devserve/config"
	"devserve/protocol"
	"encoding/json"
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

// Serve starts a new process with the given configuration.
// It auto-starts the daemon if it's not running.
func Serve(name string, port int, command, dir string) (*protocol.ServeResult, error) {
	req := &protocol.Request{
		Action: "serve",
		Args: map[string]any{
			"name":    name,
			"port":    port,
			"command": command,
			"cwd":     dir,
		},
	}

	resp, err := Send(req)
	if err != nil {
		if errors.Is(err, ErrDaemonNotRunning) {
			// Auto-start the daemon
			if startErr := StartDaemon(); startErr != nil {
				return nil, fmt.Errorf("failed to auto-start daemon: %w", startErr)
			}
			// Retry the request
			resp, err = Send(req)
		}
		if err != nil {
			return nil, err
		}
	}

	if !resp.OK {
		return nil, errors.New(resp.Error)
	}

	var result protocol.ServeResult
	if err := json.Unmarshal([]byte(resp.Data), &result); err != nil {
		return nil, fmt.Errorf("failed to parse serve response: %w", err)
	}

	return &result, nil
}

// Stop stops a running process.
func Stop(name string) error {
	req := &protocol.Request{
		Action: "stop",
		Args: map[string]any{
			"name": name,
		},
	}

	resp, err := Send(req)
	if err != nil {
		return err
	}

	if !resp.OK {
		return errors.New(resp.Error)
	}

	return nil
}

// List returns all running processes and Tailscale info.
func List() (*protocol.ListResult, error) {
	req := &protocol.Request{
		Action: "list",
	}

	resp, err := Send(req)
	if err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New(resp.Error)
	}

	var result protocol.ListResult
	if err := json.Unmarshal([]byte(resp.Data), &result); err != nil {
		return nil, fmt.Errorf("failed to parse list response: %w", err)
	}

	return &result, nil
}

// Get returns details for a single process.
func Get(name string) (*protocol.ProcessInfo, error) {
	req := &protocol.Request{
		Action: "get",
		Args: map[string]any{
			"name": name,
		},
	}

	resp, err := Send(req)
	if err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New(resp.Error)
	}

	var result protocol.ProcessInfo
	if err := json.Unmarshal([]byte(resp.Data), &result); err != nil {
		return nil, fmt.Errorf("failed to parse get response: %w", err)
	}

	return &result, nil
}

// Logs returns the last n lines of stdout and stderr for a process.
func Logs(name string, lines int) (*protocol.LogsResult, error) {
	req := &protocol.Request{
		Action: "logs",
		Args: map[string]any{
			"name":  name,
			"lines": fmt.Sprintf("%d", lines),
		},
	}

	resp, err := Send(req)
	if err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New(resp.Error)
	}

	var result protocol.LogsResult
	if err := json.Unmarshal([]byte(resp.Data), &result); err != nil {
		return nil, fmt.Errorf("failed to parse logs response: %w", err)
	}

	return &result, nil
}

// Ping checks if the daemon is running.
func Ping() error {
	req := &protocol.Request{
		Action: "ping",
	}

	resp, err := Send(req)
	if err != nil {
		return err
	}

	if !resp.OK {
		return errors.New(resp.Error)
	}

	return nil
}

// Shutdown stops the daemon.
func Shutdown() (string, error) {
	req := &protocol.Request{
		Action: "shutdown",
	}

	resp, err := Send(req)
	if err != nil {
		return "", err
	}

	if !resp.OK {
		return "", errors.New(resp.Error)
	}

	return resp.Data, nil
}
