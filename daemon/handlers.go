package daemon

import (
	"devserve/internal"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
)

func handleServe(args map[string]any) *internal.Response {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return internal.ErrResponse(fmt.Errorf("missing or invalid 'name' argument"))
	}

	if _, exists := processes[name]; exists {
		return internal.ErrResponse(fmt.Errorf("process name already in use"))
	}

	portVal, ok := args["port"]
	if !ok {
		return internal.ErrResponse(fmt.Errorf("missing 'port' argument"))
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

	command, ok := args["command"].(string)
	if !ok || command == "" {
		return internal.ErrResponse(fmt.Errorf("missing or invalid 'command' argument"))
	}

	cwd, _ := args["cwd"].(string) // optional, empty string if not provided

	log.Printf("serving %s on port %d", name, port)

	p, err := internal.CreateProcess(name, port, cwd)
	if err != nil {
		log.Printf("failed to create process %s: %s", name, err)
		return internal.ErrResponse(err)
	}

	err = p.Start(command)
	if err != nil {
		log.Printf("failed to start %s: %s", name, err)
		return internal.ErrResponse(err)
	}

	processes[p.Name] = p
	return internal.OkResponse(fmt.Sprintf("process '%s' started on port %d", name, port))
}

func handleStop(args map[string]any) *internal.Response {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return internal.ErrResponse(fmt.Errorf("missing or invalid 'name' argument"))
	}

	p, exists := processes[name]
	if !exists {
		return internal.ErrResponse(fmt.Errorf("process '%s' not found", name))
	}

	err := p.Stop()
	if err != nil {
		log.Printf("failed to stop %s: %s", name, err)
		return internal.ErrResponse(fmt.Errorf("couldn't stop process: %w", err))
	}

	delete(processes, name)
	log.Printf("stopped %s", name)
	return internal.OkResponse(fmt.Sprintf("process '%s' stopped", name))
}

func handleList(args map[string]any) *internal.Response {
	type entry struct {
		Name string `json:"name"`
		Port int    `json:"port"`
	}

	log.Printf("listed %d processes", len(processes))
	entries := make([]entry, 0, len(processes))
	for _, v := range processes {
		entries = append(entries, entry{Name: v.Name, Port: v.Port})
	}

	data, err := json.Marshal(entries)
	if err != nil {
		return internal.ErrResponse(err)
	}

	return internal.OkResponse(string(data))
}
