package daemon

import (
	"devserve/cli"
	"devserve/config"
	"devserve/process"
	"devserve/tunnel"
	"devserve/util"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"
)

func handlePing(args map[string]any) *Response {
	return OkResponse("pong")
}

func handleServe(args map[string]any) *Response {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return ErrResponse(fmt.Errorf("missing or invalid 'name' argument"))
	}

	mu.RLock()
	_, exists := processes[name]
	mu.RUnlock()
	if exists {
		return ErrResponse(fmt.Errorf("process '%s' already in use", name))
	}

	portVal, ok := args["port"]
	if !ok {
		return ErrResponse(fmt.Errorf("missing or invalid 'port' argument"))
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
			return ErrResponse(fmt.Errorf("invalid port: %w", err))
		}
	default:
		return ErrResponse(fmt.Errorf("invalid port type"))
	}

	if err := process.CheckPortInUse(port); err != nil {
		log.Printf("port %d in use: %s", port, err)
		return ErrResponse(err)
	}

	command, ok := args["command"].(string)
	if !ok || command == "" {
		return ErrResponse(fmt.Errorf("missing or invalid 'command' argument"))
	}

	cwd, _ := args["cwd"].(string) // optional, empty string if not provided

	p, err := process.CreateProcess(name, port, cwd, command)
	if err != nil {
		log.Printf("failed to create process '%s': %s", name, err)
		return ErrResponse(fmt.Errorf("failed to create process '%s': %w", name, err))
	}

	err = p.Start(command)
	if err != nil {
		log.Printf("failed to start process '%s': %s", name, err)
		return ErrResponse(fmt.Errorf("failed to start process '%s': %w", name, err))
	}

	mu.Lock()
	processes[p.Name] = p
	mu.Unlock()
	log.Printf("started '%s' on port %d", name, port)
	return OkResponse(fmt.Sprintf("process '%s' started on port %d", name, port))
}

func handleStop(args map[string]any) *Response {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return ErrResponse(fmt.Errorf("missing or invalid 'name' argument"))
	}

	mu.RLock()
	p, exists := processes[name]
	mu.RUnlock()
	if !exists {
		return ErrResponse(fmt.Errorf("process '%s' not found", name))
	}

	err := p.Stop()
	if err != nil {
		log.Printf("failed to stop process '%s': %s", name, err)
		return ErrResponse(fmt.Errorf("failed to stop process '%s': %w", name, err))
	}

	mu.Lock()
	delete(processes, name)
	mu.Unlock()
	log.Printf("stopped '%s' (port %d)", name, p.Port)
	return OkResponse(fmt.Sprintf("process '%s' stopped", name))
}

type listEntry struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

type listResponse struct {
	Processes []listEntry `json:"processes"`
	Hostname  string      `json:"hostname"`
	IP        string      `json:"ip"`
}

func handleList(args map[string]any) *Response {
	info, err := tunnel.GetTailscaleInfo(tunnel.DefaultRunner)
	if err != nil {
		return ErrResponse(err)
	}

	mu.RLock()
	entries := make([]listEntry, 0, len(processes))
	for _, v := range processes {
		entries = append(entries, listEntry{Name: v.Name, Port: v.Port})
	}
	mu.RUnlock()

	lr := listResponse{
		Processes: entries,
		Hostname:  info.Hostname,
		IP:        info.IP,
	}

	data, err := json.Marshal(lr)
	if err != nil {
		return ErrResponse(fmt.Errorf("failed to marshal process list: %w", err))
	}

	return OkResponse(string(data))
}

func handleLogs(args map[string]any) *Response {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return ErrResponse(fmt.Errorf("missing or invalid 'name' argument"))
	}

	mu.RLock()
	p, exists := processes[name]
	mu.RUnlock()
	if !exists {
		return ErrResponse(fmt.Errorf("process '%s' not found", name))
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
		return ErrResponse(fmt.Errorf("failed to read stdout log: %w", err))
	}

	stderrLines, err := util.LastNLines(stderrPath, lines)
	if err != nil {
		log.Printf("failed to read stderr log: %s", err)
		return ErrResponse(fmt.Errorf("failed to read stderr log: %w", err))
	}

	headerStyle := cli.Cyan
	stderrStyle := cli.Red

	var b strings.Builder
	b.WriteString(headerStyle.Render("─── stdout ───"))
	b.WriteString("\n")
	for _, line := range stdoutLines {
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(headerStyle.Render("─── stderr ───"))
	b.WriteString("\n")
	for _, line := range stderrLines {
		b.WriteString(stderrStyle.Render(line))
		b.WriteString("\n")
	}

	return OkResponse(strings.TrimRight(b.String(), "\n"))
}

func handleGet(args map[string]any) *Response {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return ErrResponse(fmt.Errorf("missing or invalid 'name' argument"))
	}

	mu.RLock()
	p, exists := processes[name]
	mu.RUnlock()
	if !exists {
		return ErrResponse(fmt.Errorf("process '%s' not found", name))
	}

	// Return process info as JSON
	info := map[string]any{
		"name":    p.Name,
		"port":    p.Port,
		"command": p.Command,
		"dir":     p.Dir,
	}

	data, err := json.Marshal(info)
	if err != nil {
		return ErrResponse(fmt.Errorf("failed to marshal process info: %w", err))
	}

	return OkResponse(string(data))
}
