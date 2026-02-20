package internal

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func Serve(port int, bg bool) error {
	pm, err := DetectPackageManager()
	if err != nil {
		return err
	}

	var stdout io.Writer
	var stderr io.Writer
	var fm *FileManager

	if bg {
		fm = NewGlobalFM()
		err = fm.InitDir()
		if err != nil {
			return fmt.Errorf("Failed to init directory: %w", err)
		}
		stdout, stderr, err = fm.CreateLogFiles()
		if err != nil {
			return fmt.Errorf("Failed creating std log files: %w", err)
		}
	} else {
		fm = NewLocalFM()
		stdout = os.Stdout
		stderr = os.Stderr
	}

	tm := NewTailscaleManager(stdout, stderr)

	err = tm.Start(port)
	if err != nil {
		return fmt.Errorf("failed to start Tailscale: %w", err)
	}

	command := PMToCommand[pm]
	args := strings.Split(command, " ")

	process := CreateProcess(port, pm, args[0], args[1:]...)

	process.SetOutputs(stdout, stderr)

	if bg {
		err = process.StartBG(fm)
		if err != nil {
			return fmt.Errorf("Failed to start process: %w", err)
		}
		return nil
	}

	err = process.Start()
	if err != nil {
		return fmt.Errorf("Failed to start process: %w", err)
	}

	return tm.Stop(port)
}
