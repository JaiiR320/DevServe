package tui

import (
	"devserve/daemon"
	"encoding/json"
	"errors"
	"fmt"
)

// sendRequest sends a request to the daemon, auto-starting it if needed.
func sendRequest(req *daemon.Request) (*daemon.Response, error) {
	resp, err := daemon.Send(req)
	if err == nil {
		return resp, nil
	}

	if !errors.Is(err, daemon.ErrDaemonNotRunning) {
		return resp, err
	}

	// Auto-start the daemon
	if startErr := daemon.Start(true); startErr != nil {
		return nil, fmt.Errorf("failed to auto-start daemon: %w", startErr)
	}

	return daemon.Send(req)
}

// fetchProcesses queries the daemon for the current process list and fetches
// detail for each process via the get RPC.
func fetchProcesses() ([]processRow, error) {
	resp, err := sendRequest(&daemon.Request{Action: "list"})
	if err != nil {
		return nil, fmt.Errorf("failed to send list request: %w", err)
	}
	if !resp.OK {
		return nil, errors.New(resp.Error)
	}

	type entry struct {
		Name string `json:"name"`
		Port int    `json:"port"`
	}
	type listResp struct {
		Processes []entry `json:"processes"`
		Hostname  string  `json:"hostname"`
		IP        string  `json:"ip"`
	}

	var lr listResp
	if err := json.Unmarshal([]byte(resp.Data), &lr); err != nil {
		return nil, fmt.Errorf("failed to parse list response: %w", err)
	}

	rows := make([]processRow, 0, len(lr.Processes))
	for _, e := range lr.Processes {
		row := processRow{
			Name:     e.Name,
			Port:     e.Port,
			LocalURL: fmt.Sprintf("http://localhost:%d", e.Port),
		}

		if lr.IP != "" {
			row.IPURL = fmt.Sprintf("http://%s:%d", lr.IP, e.Port)
		}
		if lr.Hostname != "" {
			row.DNSURL = fmt.Sprintf("https://%s:%d", lr.Hostname, e.Port)
		}

		// Fetch detail (command, dir) via get RPC
		detail, err := fetchDetail(e.Name)
		if err == nil {
			row.Command = detail.Command
			row.Dir = detail.Dir
		}

		rows = append(rows, row)
	}

	return rows, nil
}

// processDetail holds the extra fields returned by the get RPC.
type processDetail struct {
	Command string
	Dir     string
}

// fetchDetail queries the daemon for a single process's detail.
func fetchDetail(name string) (*processDetail, error) {
	resp, err := sendRequest(&daemon.Request{
		Action: "get",
		Args:   map[string]any{"name": name},
	})
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New(resp.Error)
	}

	var info struct {
		Command string `json:"command"`
		Dir     string `json:"dir"`
	}
	if err := json.Unmarshal([]byte(resp.Data), &info); err != nil {
		return nil, fmt.Errorf("failed to parse get response: %w", err)
	}

	return &processDetail{
		Command: info.Command,
		Dir:     info.Dir,
	}, nil
}

// stopProcess sends a stop request to the daemon for the named process.
func stopProcess(name string) error {
	resp, err := sendRequest(&daemon.Request{
		Action: "stop",
		Args:   map[string]any{"name": name},
	})
	if err != nil {
		return fmt.Errorf("failed to send stop request: %w", err)
	}
	if !resp.OK {
		return errors.New(resp.Error)
	}
	return nil
}
