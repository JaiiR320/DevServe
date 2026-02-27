package config

import (
	"os"
	"path/filepath"
	"time"
)

// Paths
var (
	DaemonDir  = "/tmp/devserve"
	Socket     = "/tmp/devserve.daemon.sock"
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
