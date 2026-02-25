package daemon

import (
	"devserve/internal"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
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
	stopChan := make(chan struct{}, 1)

	go func() {
		<-stopChan
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				fmt.Println("closing")
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
	conn, err := net.Dial("unix", internal.Socket)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write([]byte("stop"))
	if err != nil {
		return err
	}

	return nil
}

func handleConn(conn net.Conn, stop chan struct{}) {
	defer conn.Close()

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Printf("Error reading from connection: %s\n", err)
		return
	}
	data := string(buf[:n])
	fmt.Printf("Received: %s\n", data)

	if data == "stop" {
		stop <- struct{}{}
		return
	}

	params := strings.Split(data, "|")
	args := params[1:]

	switch params[0] {
	case "serve":
		err := handleServe(args)
		if err != nil {
			fmt.Println(err)
		}
	case "stop":
		err := handleStop(args)
		if err != nil {
			fmt.Print(err)
		}
	case "list":
		err := handleList(args)
		if err != nil {
			fmt.Print(err)
		}
	}
}

func Send() {

}
