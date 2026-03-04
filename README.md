# DevServe

Run dev servers and expose them over HTTPS via Tailscale. Logs go to files instead of your terminal, which is great for agentic coding workflows.

## Prerequisites

- Go 1.26+ (for installation from source)
- [Tailscale](https://tailscale.com) installed and connected to a tailnet

## Install

**Go install:**

```bash
go install github.com/jaiir320/devserve@latest
```

**Download binary:**

Grab the latest release from the [releases page](https://github.com/JaiiR320/DevServe/releases) and place it on your `PATH`.

**Build from source:**

```bash
git clone https://github.com/JaiiR320/DevServe.git
cd DevServe
go build -o devserve .
```

## Usage

```bash
# start a process (auto-starts daemon if needed)
devserve serve myapp 3000 "npm run dev"

# list running processes
devserve list

# view logs
devserve logs myapp
devserve logs myapp -n 100

# restart a process from saved config
devserve restart myapp

# stop a process
devserve stop myapp
```

Your app is available at `https://<tailnet-hostname>:3000` across your tailnet.

## TUI

Run the interactive UI:

```bash
devserve
```

**Keys:**
- `↑/↓` — navigate processes
- `enter` — start/stop selected process
- `s` — save/remove from config
- `q` — quit

The left pane shows all processes: configured (top) and ephemeral (bottom). Green = running, gray = stopped. The right pane shows details for the selected process.

## Configuration

Process configs are saved to `~/.config/devserve/config.json`.

```bash
# save a running process's config
devserve config save myapp

# start from a saved config
devserve start myapp

# list saved configs
devserve config list

# delete a saved config
devserve config delete myapp
```

## Daemon

The daemon runs in the background and manages processes over a Unix socket. It auto-starts when you run `devserve serve`, but can be managed directly:

```bash
devserve daemon start
devserve daemon stop
devserve daemon logs
```

## Logs

Process logs are per-project:

```
<project-dir>/.devserve/out.log
<project-dir>/.devserve/err.log
```

Daemon logs:

```
/tmp/devserve/out.log
```

## How It Works

A background daemon listens on a Unix socket at `/tmp/devserve.daemon.sock`. When you run `devserve serve`, it starts your command, redirects output to log files, waits for the port to be ready, then runs `tailscale serve` to expose it over HTTPS. Stopping a process kills the process tree and tears down the Tailscale proxy.
