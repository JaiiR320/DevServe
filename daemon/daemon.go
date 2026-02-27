package daemon

import (
	"devserve/cli"
	"devserve/process"
	"devserve/util"
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
	processes map[string]*process.Process
	mu        sync.RWMutex
)

func Start(background bool) error {
	if background {
		return startInBackground()
	}
	return startForeground()
}

func startInBackground() error {
	err := os.MkdirAll(util.DaemonDir, util.DirPermissions)
	if err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logFile, err := os.Create(filepath.Join(util.DaemonDir, util.DaemonLogFile))
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer logFile.Close()

	conn, err := net.Dial("unix", util.Socket)
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

	var startErr error
	cli.Spin("Starting daemon...", func() {
		err = cmd.Start()
		if err != nil {
			startErr = fmt.Errorf("failed to start daemon: %w", err)
			return
		}

		// Wait a moment and verify daemon started with ping
		time.Sleep(util.DaemonStartDelay)
		pingReq := &Request{Action: "ping"}
		resp, err := Send(pingReq)
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

	fmt.Println(cli.Success("daemon started") + " " + cli.Info("logs: "+filepath.Join(util.DaemonDir, util.DaemonLogFile)))
	return nil
}

func startForeground() error {
	cli.InitLogger()
	processes = make(map[string]*process.Process)
	conn, err := net.Dial("unix", util.Socket)
	if err == nil {
		conn.Close()
		return errors.New("daemon is already running")
	}
	os.Remove(util.Socket)

	listener, err := net.Listen("unix", util.Socket)
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
		go HandleConn(conn, stopChan)
	}

	// Stop all running child processes before exiting
	failed := stopAllProcesses(util.ShutdownTimeout)
	if len(failed) > 0 {
		log.Printf("failed to stop processes on ports: %s", strings.Join(failed, ", "))
	}

	os.Remove(util.Socket)
	return nil
}

func Stop() (string, error) {
	req := &Request{Action: "shutdown"}
	resp, err := Send(req)
	if err != nil {
		return "", fmt.Errorf("failed to send shutdown request: %w", err)
	}
	if !resp.OK {
		return "", fmt.Errorf("failed to shutdown daemon: %s", resp.Error)
	}
	return resp.Data, nil
}

// StopAllProcesses is exported for testing.
func StopAllProcesses(timeout time.Duration) []string {
	return stopAllProcesses(timeout)
}

// stopAllProcesses stops all running child processes with retry logic.
func stopAllProcesses(timeout time.Duration) []string {
	mu.Lock()
	snapshot := make(map[string]*process.Process, len(processes))
	for k, v := range processes {
		snapshot[k] = v
	}
	mu.Unlock()

	if len(snapshot) == 0 {
		return nil
	}

	const maxRetries = 3

	type result struct {
		name    string
		port    int
		err     error
		attempt int
	}

	results := make(chan result, len(snapshot)*maxRetries)
	for name, p := range snapshot {
		go func(name string, p *process.Process) {
			err := p.Stop()
			results <- result{name: name, port: p.Port, err: err, attempt: 1}
		}(name, p)
	}

	var failed []string
	timer := time.After(timeout)
	remaining := len(snapshot)

	for remaining > 0 {
		select {
		case r := <-results:
			if r.err != nil {
				log.Printf("failed to stop %s (port %d) attempt %d: %s", r.name, r.port, r.attempt, r.err)
				if r.attempt < maxRetries {
					p := snapshot[r.name]
					go func(name string, p *process.Process, attempt int) {
						time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
						err := p.Stop()
						results <- result{name: name, port: p.Port, err: err, attempt: attempt + 1}
					}(r.name, p, r.attempt)
				} else {
					remaining--
					failed = append(failed, fmt.Sprintf("%d", r.port))
				}
			} else {
				remaining--
				log.Printf("stopped %s (port %d)", r.name, r.port)
				mu.Lock()
				delete(processes, r.name)
				mu.Unlock()
			}
		case <-timer:
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

func HandleConn(conn net.Conn, stop chan struct{}) {
	defer conn.Close()

	req, err := ReadRequest(conn)
	if err != nil {
		log.Printf("failed to read request: %s", err)
		SendResponse(conn, ErrResponse(err))
		return
	}

	if req.Action == "shutdown" {
		log.Println("shutdown requested, stopping all processes")
		failed := stopAllProcesses(util.ShutdownTimeout)
		if len(failed) > 0 {
			msg := fmt.Sprintf("daemon stopping, failed to stop ports: %s", strings.Join(failed, ", "))
			SendResponse(conn, OkResponse(msg))
		} else {
			SendResponse(conn, OkResponse("daemon stopped, all processes terminated"))
		}
		stop <- struct{}{}
		return
	}

	var resp *Response
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
		resp = ErrResponse(fmt.Errorf("unknown action '%s'", req.Action))
	}

	SendResponse(conn, resp)
}
