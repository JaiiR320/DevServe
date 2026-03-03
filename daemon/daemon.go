package daemon

import (
	"devserve/config"
	"devserve/process"
	"devserve/protocol"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	processes map[string]*process.Process
	mu        sync.RWMutex
)

// Run starts the daemon in foreground mode.
// It listens on the Unix socket, accepts connections, and handles requests
// until a shutdown signal is received.
func Run() error {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime)
	processes = make(map[string]*process.Process)
	conn, err := net.Dial("unix", config.Socket)
	if err == nil {
		conn.Close()
		return errors.New("daemon is already running")
	}
	os.Remove(config.Socket)

	listener, err := net.Listen("unix", config.Socket)
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
	failed := stopAllProcesses(config.ShutdownTimeout)
	if len(failed) > 0 {
		log.Printf("failed to stop processes on ports: %s", strings.Join(failed, ", "))
	}

	os.Remove(config.Socket)
	return nil
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

// handleConn handles a single connection to the daemon.
func handleConn(conn net.Conn, stop chan struct{}) {
	defer conn.Close()

	req, err := protocol.ReadRequest(conn)
	if err != nil {
		log.Printf("failed to read request: %s", err)
		protocol.SendResponse(conn, protocol.ErrResponse(err))
		return
	}

	if req.Action == "shutdown" {
		log.Println("shutdown requested, stopping all processes")
		failed := stopAllProcesses(config.ShutdownTimeout)
		if len(failed) > 0 {
			msg := fmt.Sprintf("daemon stopping, failed to stop ports: %s", strings.Join(failed, ", "))
			protocol.SendResponse(conn, protocol.OkResponse(msg))
		} else {
			protocol.SendResponse(conn, protocol.OkResponse("daemon stopped, all processes terminated"))
		}
		stop <- struct{}{}
		return
	}

	var resp *protocol.Response
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
	case "get":
		resp = handleGet(req.Args)
	default:
		resp = protocol.ErrResponse(fmt.Errorf("unknown action '%s'", req.Action))
	}

	protocol.SendResponse(conn, resp)
}
