package tui

import (
	"devserve/client"
	"devserve/config"
	"fmt"
	"sort"
)

// fetchItems queries the daemon and config to build a unified list of processes.
// Configured items come first, followed by ephemeral (running but not configured) items.
func fetchItems() ([]listItem, error) {
	// Fetch running processes from daemon
	lr, err := client.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list: %w", err)
	}

	// Build map of running processes
	runningProcs := make(map[string]processInfo, len(lr.Processes))
	for _, e := range lr.Processes {
		info := processInfo{
			Name:     e.Name,
			Port:     e.Port,
			LocalURL: fmt.Sprintf("http://localhost:%d", e.Port),
		}
		if lr.IP != "" {
			info.IPURL = fmt.Sprintf("http://%s:%d", lr.IP, e.Port)
		}
		if lr.Hostname != "" {
			info.DNSURL = fmt.Sprintf("https://%s:%d", lr.Hostname, e.Port)
		}

		// Fetch detail (command, dir) via get RPC
		detail, err := client.Get(e.Name)
		if err == nil {
			info.Command = detail.Command
			info.Dir = detail.Dir
		}

		runningProcs[e.Name] = info
	}

	// Load configs
	configs, err := config.LoadConfigs(config.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load configs: %w", err)
	}

	// Build unified list
	var items []listItem
	configuredNames := make(map[string]bool)

	// First: configured items (whether running or not)
	for _, cfg := range configs {
		item := listItem{
			Name:       cfg.Name,
			Port:       cfg.Port,
			Command:    cfg.Command,
			Dir:        cfg.Directory,
			Configured: true,
		}

		// Check if running
		if proc, ok := runningProcs[cfg.Name]; ok {
			item.Running = true
			item.LocalURL = proc.LocalURL
			item.IPURL = proc.IPURL
			item.DNSURL = proc.DNSURL
			// Update with live command/dir from running process
			item.Command = proc.Command
			item.Dir = proc.Dir
		}

		items = append(items, item)
		configuredNames[cfg.Name] = true
	}

	// Sort configured items by port
	sort.Slice(items, func(i, j int) bool {
		return items[i].Port < items[j].Port
	})

	// Second: ephemeral items (running but not configured)
	var ephemeral []listItem
	for name, proc := range runningProcs {
		if !configuredNames[name] {
			ephemeral = append(ephemeral, listItem{
				Name:       proc.Name,
				Port:       proc.Port,
				Command:    proc.Command,
				Dir:        proc.Dir,
				Running:    true,
				Configured: false,
				LocalURL:   proc.LocalURL,
				IPURL:      proc.IPURL,
				DNSURL:     proc.DNSURL,
			})
		}
	}

	// Sort ephemeral items by port
	sort.Slice(ephemeral, func(i, j int) bool {
		return ephemeral[i].Port < ephemeral[j].Port
	})

	// Combine: configured first, then ephemeral
	items = append(items, ephemeral...)

	return items, nil
}

// processInfo holds temporary process data during fetch
type processInfo struct {
	Name     string
	Port     int
	Command  string
	Dir      string
	LocalURL string
	IPURL    string
	DNSURL   string
}

// stopProcess sends a stop request to the daemon for the named process.
func stopProcess(name string) error {
	return client.Stop(name)
}

// startItem starts a configured process.
func startItem(item listItem) error {
	_, err := client.Serve(item.Name, item.Port, item.Command, item.Dir)
	return err
}

// saveToConfig saves an ephemeral process to the config file.
func saveToConfig(item listItem) error {
	newConfig := config.ProcessConfig{
		Name:      item.Name,
		Port:      item.Port,
		Command:   item.Command,
		Directory: item.Dir,
	}
	return config.SaveConfig(config.ConfigFile, newConfig)
}

// removeFromConfig removes a process from the config file.
func removeFromConfig(name string) error {
	return config.DeleteConfig(config.ConfigFile, name)
}
