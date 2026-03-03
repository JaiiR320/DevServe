package config

import (
	"os"
	"path/filepath"
	"time"
)

// socketPath returns the appropriate socket path based on XDG_RUNTIME_DIR.
func socketPath() string {
	if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
		return filepath.Join(runtimeDir, "devserve.daemon.sock")
	}
	return "/tmp/devserve.daemon.sock"
}

// daemonDir returns the appropriate daemon directory based on XDG_RUNTIME_DIR.
func daemonDir() string {
	if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
		return filepath.Join(runtimeDir, "devserve")
	}
	return "/tmp/devserve"
}

// Paths
var (
	DaemonDir  = daemonDir()
	Socket     = socketPath()
	ConfigDir  = filepath.Join(os.Getenv("HOME"), ".config", "devserve")
	ConfigFile = filepath.Join(ConfigDir, "config.json")
)

const (
	DaemonLogFile    = "out.log"
	ProcessLogDir    = ".devserve"
	ProcessStdoutLog = "out.log"
	ProcessStderrLog = "err.log"
)

// Timeouts
const (
	PortWaitTimeout  = 15 * time.Second
	StopGracePeriod  = 5 * time.Second
	PortDialTimeout  = 500 * time.Millisecond
	PortPollInterval = 500 * time.Millisecond
	DaemonStartDelay = 100 * time.Millisecond
	ShutdownTimeout  = 15 * time.Second
)

// Permissions
const DirPermissions = os.FileMode(0755)
