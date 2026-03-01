# DevServe

Run dev servers and expose them over HTTPS via Tailscale. Logs go to files instead of your terminal, which is great for agentic coding workflows.

## Usage

```bash
# start a process (auto-starts daemon if needed)
devserve serve myapp 3000 "npm run dev"

# shorthand — same thing
devserve myapp 3000 "npm run dev"

# list running processes
devserve list

# view logs
devserve logs myapp
devserve logs myapp -n 100

# stop a process
devserve stop myapp
```

Your app is available at `https://<tailnet-hostname>:3000` across your tailnet.

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

Requires [Tailscale](https://tailscale.com) installed and connected to a tailnet.
