package daemon

import (
	"devserve/cli"
	"devserve/internal"
	"devserve/util"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"
)

func handlePing(args map[string]any) *internal.Response {
	return internal.OkResponse("pong")
}

func handleServe(args map[string]any) *internal.Response {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return internal.ErrResponse(fmt.Errorf("missing or invalid 'name' argument"))
	}

	mu.RLock()
	_, exists := processes[name]
	mu.RUnlock()
	if exists {
		return internal.ErrResponse(fmt.Errorf("process '%s' already in use", name))
	}

	portVal, ok := args["port"]
	if !ok {
		return internal.ErrResponse(fmt.Errorf("missing or invalid 'port' argument"))
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
			return internal.ErrResponse(fmt.Errorf("invalid port: %w", err))
		}
	default:
		return internal.ErrResponse(fmt.Errorf("invalid port type"))
	}

	if err := internal.CheckPortInUse(port); err != nil {
		log.Printf("port %d in use: %s", port, err)
		return internal.ErrResponse(err)
	}

	command, ok := args["command"].(string)
	if !ok || command == "" {
		return internal.ErrResponse(fmt.Errorf("missing or invalid 'command' argument"))
	}

	cwd, _ := args["cwd"].(string) // optional, empty string if not provided

	p, err := internal.CreateProcess(name, port, cwd)
	if err != nil {
		log.Printf("failed to create process '%s': %s", name, err)
		return internal.ErrResponse(fmt.Errorf("failed to create process '%s': %w", name, err))
	}

	err = p.Start(command)
	if err != nil {
		log.Printf("failed to start process '%s': %s", name, err)
		return internal.ErrResponse(fmt.Errorf("failed to start process '%s': %w", name, err))
	}

	mu.Lock()
	processes[p.Name] = p
	mu.Unlock()
	log.Printf("started '%s' on port %d", name, port)
	return internal.OkResponse(fmt.Sprintf("process '%s' started on port %d", name, port))
}

func handleStop(args map[string]any) *internal.Response {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return internal.ErrResponse(fmt.Errorf("missing or invalid 'name' argument"))
	}

	mu.RLock()
	p, exists := processes[name]
	mu.RUnlock()
	if !exists {
		return internal.ErrResponse(fmt.Errorf("process '%s' not found", name))
	}

	err := p.Stop()
	if err != nil {
		log.Printf("failed to stop process '%s': %s", name, err)
		return internal.ErrResponse(fmt.Errorf("failed to stop process '%s': %w", name, err))
	}

	mu.Lock()
	delete(processes, name)
	mu.Unlock()
	log.Printf("stopped '%s' (port %d)", name, p.Port)
	return internal.OkResponse(fmt.Sprintf("process '%s' stopped", name))
}

// tsRunner is the function used to fetch tailscale status JSON.
// It can be overridden in tests.
var tsRunner internal.CommandRunner = internal.DefaultTailscaleRunner

type listEntry struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

type listResponse struct {
	Processes []listEntry `json:"processes"`
	Hostname  string      `json:"hostname"`
	IP        string      `json:"ip"`
}

func handleList(args map[string]any) *internal.Response {
	info, err := internal.GetTailscaleInfo(tsRunner)
	if err != nil {
		return internal.ErrResponse(err)
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
		return internal.ErrResponse(fmt.Errorf("failed to marshal process list: %w", err))
	}

	return internal.OkResponse(string(data))
}

// ResetProcesses clears the processes map for test isolation.
func ResetProcesses() {
	mu.Lock()
	processes = make(map[string]*internal.Process)
	mu.Unlock()
}

// SetProcess adds a process to the map (for test setup).
func SetProcess(name string, p *internal.Process) {
	mu.Lock()
	if processes == nil {
		processes = make(map[string]*internal.Process)
	}
	processes[name] = p
	mu.Unlock()
}

// GetProcesses returns a copy of the processes map (for test assertions).
func GetProcesses() map[string]*internal.Process {
	mu.RLock()
	defer mu.RUnlock()
	cp := make(map[string]*internal.Process, len(processes))
	for k, v := range processes {
		cp[k] = v
	}
	return cp
}

func handleLogs(args map[string]any) *internal.Response {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return internal.ErrResponse(fmt.Errorf("missing or invalid 'name' argument"))
	}

	mu.RLock()
	p, exists := processes[name]
	mu.RUnlock()
	if !exists {
		return internal.ErrResponse(fmt.Errorf("process '%s' not found", name))
	}

	lines := 50
	if linesStr, ok := args["lines"].(string); ok {
		if n, err := strconv.Atoi(linesStr); err == nil && n > 0 {
			lines = n
		}
	}

	stdoutPath := filepath.Join(p.Dir, util.ProcessLogDir, util.ProcessStdoutLog)
	stderrPath := filepath.Join(p.Dir, util.ProcessLogDir, util.ProcessStderrLog)

	stdoutLines, err := util.LastNLines(stdoutPath, lines)
	if err != nil {
		log.Printf("failed to read stdout log: %s", err)
		return internal.ErrResponse(fmt.Errorf("failed to read stdout log: %w", err))
	}

	stderrLines, err := util.LastNLines(stderrPath, lines)
	if err != nil {
		log.Printf("failed to read stderr log: %s", err)
		return internal.ErrResponse(fmt.Errorf("failed to read stderr log: %w", err))
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

	return internal.OkResponse(strings.TrimRight(b.String(), "\n"))
}
