# DevServe

Run dev servers and expose them over HTTPS via Tailscale. Logs go to files instead of your terminal, which is great for agentic coding workflows.

## Install

```bash
go build -o devserve .
```

Requires [Tailscale](https://tailscale.com) installed and connected to a tailnet.

## Usage

```bash
# start the daemon
devserve daemon start

# serve a project (name, port, command)
devserve serve myapp 3000 "npm run dev"

# list running processes
devserve list

# stop a process
devserve stop myapp

# stop the daemon
devserve daemon stop
```

Your app is now available at `https://<your-tailnet-magicdns>:3000` across your tailnet.

## Logs

Stdout and stderr are redirected to files so they don't clutter your terminal:

```
/.devserve/out.log
/.devserve/err.log
```

## How it works

A background daemon manages processes over a Unix socket. When you run `devserve serve`, it starts your command, redirects its output to log files, and runs `tailscale serve` to expose it over HTTPS. Stopping a process kills the process tree and tears down the Tailscale proxy.
