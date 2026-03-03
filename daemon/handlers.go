package daemon

import (
	"devserve/config"
	"devserve/process"
	"devserve/protocol"
	"devserve/tunnel"
	"devserve/util"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
)

func handlePing(args map[string]any) *protocol.Response {
	return protocol.OkResponse("pong")
}

func handleServe(args map[string]any) *protocol.Response {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return protocol.ErrResponse(fmt.Errorf("missing or invalid 'name' argument"))
	}

	mu.RLock()
	_, exists := processes[name]
	mu.RUnlock()
	if exists {
		return protocol.ErrResponse(fmt.Errorf("process '%s' already in use", name))
	}

	portVal, ok := args["port"]
	if !ok {
		return protocol.ErrResponse(fmt.Errorf("missing or invalid 'port' argument"))
	}
	// JSON numbers decode as float64
	var port int
	switch v := portVal.(type) {
	case float64:
		port = int(v)
	case string:
		var err error
		port, err = strconv.Atoi(v)
		if err != nil {
			return protocol.ErrResponse(fmt.Errorf("invalid port: %w", err))
		}
	default:
		return protocol.ErrResponse(fmt.Errorf("invalid port type"))
	}

	if err := process.CheckPortInUse(port); err != nil {
		log.Printf("port %d in use: %s", port, err)
		return protocol.ErrResponse(err)
	}

	command, ok := args["command"].(string)
	if !ok || command == "" {
		return protocol.ErrResponse(fmt.Errorf("missing or invalid 'command' argument"))
	}

	cwd, _ := args["cwd"].(string) // optional, empty string if not provided

	p, err := process.CreateProcess(name, port, cwd, command)
	if err != nil {
		log.Printf("failed to create process '%s': %s", name, err)
		return protocol.ErrResponse(fmt.Errorf("failed to create process '%s': %w", name, err))
	}

	err = p.Start(command)
	if err != nil {
		log.Printf("failed to start process '%s': %s", name, err)
		return protocol.ErrResponse(fmt.Errorf("failed to start process '%s': %w", name, err))
	}

	mu.Lock()
	processes[p.Name] = p
	mu.Unlock()
	log.Printf("started '%s' on port %d", name, port)

	sr := protocol.ServeResult{Name: name, Port: port}
	if info, err := tunnel.GetTailscaleInfo(tunnel.DefaultRunner); err == nil {
		sr.Hostname = info.Hostname
		sr.IP = info.IP
	} else {
		log.Printf("failed to get tailscale info: %s", err)
	}

	data, err := json.Marshal(sr)
	if err != nil {
		return protocol.OkResponse(fmt.Sprintf("process '%s' started on port %d", name, port))
	}
	return protocol.OkResponse(string(data))
}

func handleStop(args map[string]any) *protocol.Response {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return protocol.ErrResponse(fmt.Errorf("missing or invalid 'name' argument"))
	}

	mu.RLock()
	p, exists := processes[name]
	mu.RUnlock()
	if !exists {
		return protocol.ErrResponse(fmt.Errorf("process '%s' not found", name))
	}

	err := p.Stop()
	if err != nil {
		log.Printf("failed to stop process '%s': %s", name, err)
		return protocol.ErrResponse(fmt.Errorf("failed to stop process '%s': %w", name, err))
	}

	mu.Lock()
	delete(processes, name)
	mu.Unlock()
	log.Printf("stopped '%s' (port %d)", name, p.Port)
	return protocol.OkResponse(fmt.Sprintf("process '%s' stopped", name))
}

func handleList(args map[string]any) *protocol.Response {
	info, err := tunnel.GetTailscaleInfo(tunnel.DefaultRunner)
	if err != nil {
		return protocol.ErrResponse(err)
	}

	mu.RLock()
	entries := make([]protocol.ListEntry, 0, len(processes))
	for _, v := range processes {
		entries = append(entries, protocol.ListEntry{Name: v.Name, Port: v.Port})
	}
	mu.RUnlock()

	lr := protocol.ListResult{
		Processes: entries,
		Hostname:  info.Hostname,
		IP:        info.IP,
	}

	data, err := json.Marshal(lr)
	if err != nil {
		return protocol.ErrResponse(fmt.Errorf("failed to marshal process list: %w", err))
	}

	return protocol.OkResponse(string(data))
}

func handleLogs(args map[string]any) *protocol.Response {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return protocol.ErrResponse(fmt.Errorf("missing or invalid 'name' argument"))
	}

	mu.RLock()
	p, exists := processes[name]
	mu.RUnlock()
	if !exists {
		return protocol.ErrResponse(fmt.Errorf("process '%s' not found", name))
	}

	lines := 50
	if linesStr, ok := args["lines"].(string); ok {
		if n, err := strconv.Atoi(linesStr); err == nil && n > 0 {
			lines = n
		}
	}

	stdoutPath := filepath.Join(p.Dir, config.ProcessLogDir, config.ProcessStdoutLog)
	stderrPath := filepath.Join(p.Dir, config.ProcessLogDir, config.ProcessStderrLog)

	stdoutLines, err := util.LastNLines(stdoutPath, lines)
	if err != nil {
		log.Printf("failed to read stdout log: %s", err)
		return protocol.ErrResponse(fmt.Errorf("failed to read stdout log: %w", err))
	}

	stderrLines, err := util.LastNLines(stderrPath, lines)
	if err != nil {
		log.Printf("failed to read stderr log: %s", err)
		return protocol.ErrResponse(fmt.Errorf("failed to read stderr log: %w", err))
	}

	result := protocol.LogsResult{
		Stdout: stdoutLines,
		Stderr: stderrLines,
	}

	data, err := json.Marshal(result)
	if err != nil {
		return protocol.ErrResponse(fmt.Errorf("failed to marshal logs: %w", err))
	}

	return protocol.OkResponse(string(data))
}

func handleGet(args map[string]any) *protocol.Response {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return protocol.ErrResponse(fmt.Errorf("missing or invalid 'name' argument"))
	}

	mu.RLock()
	p, exists := processes[name]
	mu.RUnlock()
	if !exists {
		return protocol.ErrResponse(fmt.Errorf("process '%s' not found", name))
	}

	// Return process info as structured data
	info := protocol.ProcessInfo{
		Name:    p.Name,
		Port:    p.Port,
		Command: p.Command,
		Dir:     p.Dir,
	}

	data, err := json.Marshal(info)
	if err != nil {
		return protocol.ErrResponse(fmt.Errorf("failed to marshal process info: %w", err))
	}

	return protocol.OkResponse(string(data))
}
