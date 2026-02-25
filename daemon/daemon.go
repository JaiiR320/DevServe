package daemon

import (
	"devserve/internal"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
)

var processes map[string]*internal.Process

func Start() error {
	processes = make(map[string]*internal.Process)
	conn, err := net.Dial("unix", internal.Socket)
	if err == nil {
		conn.Close()
		return errors.New("Another daemon is already running")
	}
	os.Remove(internal.Socket)

	listener, err := net.Listen("unix", internal.Socket)
	if err != nil {
		return err
	}
	defer listener.Close()
	log.Println("daemon started")
	stopChan := make(chan struct{}, 1)

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
	os.Remove(internal.Socket)
	return nil
}

func Stop() error {
	req := &internal.Request{Action: "shutdown"}
	resp, err := internal.Send(req)
	if err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("daemon shutdown failed: %s", resp.Error)
	}
	return nil
}

func handleConn(conn net.Conn, stop chan struct{}) {
	defer conn.Close()

	req, err := internal.ReadRequest(conn)
	if err != nil {
		log.Println("error reading request:", err)
		internal.SendResponse(conn, internal.ErrResponse(err))
		return
	}

	log.Println("request:", req.Action)

	if req.Action == "shutdown" {
		internal.SendResponse(conn, internal.OkResponse("daemon stopping"))
		stop <- struct{}{}
		return
	}

	var resp *internal.Response
	switch req.Action {
	case "serve":
		resp = handleServe(req.Args)
	case "stop":
		resp = handleStop(req.Args)
	case "list":
		resp = handleList(req.Args)
	default:
		resp = internal.ErrResponse(fmt.Errorf("unknown action: %s", req.Action))
	}

	internal.SendResponse(conn, resp)
}
