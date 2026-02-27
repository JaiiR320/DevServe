package daemon

import (
	"devserve/internal"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	processes map[string]*internal.Process
	mu        sync.RWMutex
)

func Start(background bool) error {
	if background {
		return startInBackground()
	}
	return startForeground()
}

func startInBackground() error {
	// Create log directory
	err := os.MkdirAll(internal.DaemonDir, internal.DirPermissions)
	if err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file (truncate if exists)
	logFile, err := os.Create(filepath.Join(internal.DaemonDir, internal.DaemonLogFile))
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer logFile.Close()

	// Check if daemon already running
	conn, err := net.Dial("unix", internal.Socket)
	if err == nil {
		conn.Close()
		return errors.New("daemon is already running")
	}

	// Get the current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Fork process
	cmd := exec.Command(execPath, "daemon", "start", "--foreground")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	var startErr error
	internal.Spin("Starting daemon...", func() {
		err = cmd.Start()
		if err != nil {
			startErr = fmt.Errorf("failed to start daemon: %w", err)
			return
		}

		// Wait a moment and verify daemon started with ping
		time.Sleep(internal.DaemonStartDelay)
		pingReq := &internal.Request{Action: "ping"}
		resp, err := internal.Send(pingReq)
		if err != nil {
			startErr = fmt.Errorf("failed to start daemon: %w", err)
			return
		}
		if !resp.OK || resp.Data != "pong" {
			startErr = errors.New("daemon health check failed")
			return
		}
	})
	if startErr != nil {
		return startErr
	}

	fmt.Println(internal.Success("daemon started") + " " + internal.Info("logs: "+filepath.Join(internal.DaemonDir, internal.DaemonLogFile)))
	return nil
}

func startForeground() error {
	internal.InitLogger()
	processes = make(map[string]*internal.Process)
	conn, err := net.Dial("unix", internal.Socket)
	if err == nil {
		conn.Close()
		return errors.New("daemon is already running")
	}
	os.Remove(internal.Socket)

	listener, err := net.Listen("unix", internal.Socket)
	if err != nil {
		return fmt.Errorf("failed to listen on socket: %w", err)
	}
	defer listener.Close()
	log.Println("daemon started")
	stopChan := make(chan struct{}, 1)

	// Handle OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Printf("received signal: %s", sig)
		stopChan <- struct{}{}
	}()

	go func() {
		<-stopChan
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				log.Println("daemon shutting down")
				break
			}
			continue
		}
		go handleConn(conn, stopChan)
	}

	// Stop all running child processes before exiting
	failed := stopAllProcesses(internal.ShutdownTimeout)
	if len(failed) > 0 {
		log.Printf("failed to stop processes on ports: %s", strings.Join(failed, ", "))
	}

	os.Remove(internal.Socket)
	return nil
}

func Stop() (string, error) {
	req := &internal.Request{Action: "shutdown"}
	resp, err := internal.Send(req)
	if err != nil {
		return "", fmt.Errorf("failed to send shutdown request: %w", err)
	}
	if !resp.OK {
		return "", fmt.Errorf("failed to shutdown daemon: %s", resp.Error)
	}
	return resp.Data, nil
}

// stopAllProcesses stops all running child processes concurrently.
// Returns a list of port strings for processes that failed to stop within the timeout.
func stopAllProcesses(timeout time.Duration) []string {
	mu.Lock()
	snapshot := make(map[string]*internal.Process, len(processes))
	for k, v := range processes {
		snapshot[k] = v
	}
	mu.Unlock()

	if len(snapshot) == 0 {
		return nil
	}

	type result struct {
		name string
		port int
		err  error
	}

	results := make(chan result, len(snapshot))
	for name, p := range snapshot {
		go func(name string, p *internal.Process) {
			err := p.Stop()
			results <- result{name: name, port: p.Port, err: err}
		}(name, p)
	}

	var failed []string
	timer := time.After(timeout)
	remaining := len(snapshot)

	for remaining > 0 {
		select {
		case r := <-results:
			remaining--
			if r.err != nil {
				log.Printf("failed to stop %s (port %d): %s", r.name, r.port, r.err)
				failed = append(failed, fmt.Sprintf("%d", r.port))
			} else {
				log.Printf("stopped %s (port %d)", r.name, r.port)
				mu.Lock()
				delete(processes, r.name)
				mu.Unlock()
			}
		case <-timer:
			// Collect any remaining processes as failed
			mu.RLock()
			for _, p := range processes {
				failed = append(failed, fmt.Sprintf("%d", p.Port))
			}
			mu.RUnlock()
			return failed
		}
	}

	return failed
}

func handleConn(conn net.Conn, stop chan struct{}) {
	defer conn.Close()

	req, err := internal.ReadRequest(conn)
	if err != nil {
		log.Printf("failed to read request: %s", err)
		internal.SendResponse(conn, internal.ErrResponse(err))
		return
	}

	if req.Action == "shutdown" {
		log.Println("shutdown requested, stopping all processes")
		failed := stopAllProcesses(internal.ShutdownTimeout)
		if len(failed) > 0 {
			msg := fmt.Sprintf("daemon stopping, failed to stop ports: %s", strings.Join(failed, ", "))
			internal.SendResponse(conn, internal.OkResponse(msg))
		} else {
			internal.SendResponse(conn, internal.OkResponse("daemon stopped, all processes terminated"))
		}
		stop <- struct{}{}
		return
	}

	var resp *internal.Response
	switch req.Action {
	case "ping":
		resp = handlePing(req.Args)
	case "serve":
		resp = handleServe(req.Args)
	case "stop":
		resp = handleStop(req.Args)
	case "list":
		resp = handleList(req.Args)
	case "logs":
		resp = handleLogs(req.Args)
	default:
		resp = internal.ErrResponse(fmt.Errorf("unknown action '%s'", req.Action))
	}

	internal.SendResponse(conn, resp)
}
