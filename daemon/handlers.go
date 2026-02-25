package daemon

import (
	"devserve/internal"
	"fmt"
	"strconv"
)

func handleServe(args []string) error {
	name := args[0]
	if _, ok := processes[name]; ok {
		return fmt.Errorf("process name already in use")
	}

	port, err := strconv.Atoi(args[1])
	if err != nil {
		return err
	}
	command := args[2]
	p, err := internal.CreateProcess(name, port)
	if err != nil {
		return err
	}
	err = p.Start(command)
	if err != nil {
		return err
	}
	processes[p.Name] = p
	return nil
}

func handleStop(args []string) error {
	name := args[0]
	if p, ok := processes[name]; ok {
		err := p.Stop()
		if err != nil {
			return fmt.Errorf("couldn't stop process: %w", err)
		}
	}
	return nil
}

func handleList(args []string) error {
	for _, v := range processes {
		portstr := strconv.Itoa(v.Port)
		fmt.Println(v.Name + " | " + portstr)
	}
	return nil
}
