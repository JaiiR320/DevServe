package internal

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Start detects the package manager, starts the dev server, and serves it over Tailscale.
func Start(port int, bg bool) error {
	pm, err := DetectPackageManager()
	if err != nil {
		return err
	}

	var stdout io.Writer
	var stderr io.Writer
	var fm *FileManager

	if bg {
		fm = NewGlobalFM()
		stdout, stderr, err = fm.CreateLogFiles()
		if err != nil {
			return fmt.Errorf("failed creating std log files: %w", err)
		}
	} else {
		fm = NewLocalFM()
		stdout = os.Stdout
		stderr = os.Stderr
	}

	command := PMToCommand[pm]
	args := strings.Split(command, " ")

	process := CreateProcess(port, pm, args[0], args[1:]...)

	process.SetOutputs(stdout, stderr)

	err = process.Start(fm)
	if err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	tm := NewTailscaleManager(stdout, stderr)

	err = tm.Start(port)
	if err != nil {
		return fmt.Errorf("failed to start Tailscale: %w", err)
	}

	if bg {
		return nil
	}

	err = process.Wait()
	if err != nil {
		return fmt.Errorf("failed on wait: %w", err)
	}
	err = tm.Stop(port)
	if err != nil {
		return fmt.Errorf("failed to stop tailscale: %w", err)
	}
	return nil
}
