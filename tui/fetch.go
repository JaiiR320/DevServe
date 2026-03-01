package tui

import (
	"devserve/config"
	"devserve/daemon"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
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

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Port < rows[j].Port
	})

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

// fetchConfigs loads saved configurations and cross-references with running
// processes to set the Running flag.
func fetchConfigs(processes []processRow) ([]configRow, error) {
	configs, err := config.LoadConfigs(config.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load configs: %w", err)
	}

	// Build a set of running process names for fast lookup
	running := make(map[string]bool, len(processes))
	for _, p := range processes {
		running[p.Name] = true
	}

	rows := make([]configRow, len(configs))
	for i, c := range configs {
		rows[i] = configRow{
			Name:    c.Name,
			Port:    c.Port,
			Command: c.Command,
			Dir:     c.Directory,
			Running: running[c.Name],
		}
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Port < rows[j].Port
	})

	return rows, nil
}

// startProcess sends a serve request to the daemon to start a saved config.
func startProcess(cfg configRow) error {
	resp, err := sendRequest(&daemon.Request{
		Action: "serve",
		Args: map[string]any{
			"name":    cfg.Name,
			"port":    cfg.Port,
			"command": cfg.Command,
			"cwd":     cfg.Dir,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send serve request: %w", err)
	}
	if !resp.OK {
		return errors.New(resp.Error)
	}
	return nil
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
