# DevServe Architecture

## Transport
Unix socket: `/tmp/devserve.daemon.sock` (hardcoded `internal/config.go`). One connection per command. No HTTP layer.

## Protocol (`internal/protocol.go`)
JSON over socket, newline-delimited.

Request: `{ "action": "<str>", "args": { ... } }`
Response: `{ "ok": bool, "data": "<str>", "error": "<str>" }`

## Actions
- `serve` — args: name, port, command, cwd → starts process + tailscale proxy
- `stop` — args: name → SIGTERM/SIGKILL + tailscale teardown
- `list` — no args → JSON array of {name, port}
- `shutdown` — no args → stops all processes, exits daemon

## Daemon Lifecycle (`daemon/daemon.go`)
Start: check no existing daemon → listen on socket → accept loop → goroutine per conn.
Stop: `shutdown` action or SIGINT/SIGTERM → stop all processes (15s timeout) → remove socket.

## Client (`internal/client.go`)
`Send()`: dial socket → encode request → decode response → close. CLI checks `resp.OK`.

## Auth
None. Unix socket file permissions only.

## Key Files
- `internal/config.go` — socket path
- `internal/protocol.go` — request/response types + encoding
- `internal/client.go` — client Send()
- `internal/process.go` — process start/stop + tailscale
- `daemon/daemon.go` — daemon loop + shutdown
- `daemon/handlers.go` — action handlers
- `cmd/` — CLI commands (serve, stop, list, daemon)

## Learnings

### Error handling conventions
Wrap errors with `fmt.Errorf("failed to <verb> <noun>: %w", err)`. Always single-quote user-provided names in errors (e.g. `"process '%s' not found"`). Use `"failed to <verb>"` prefix for error log messages. Keep log-then-return pattern in handlers. Validation errors use `"missing or invalid '<field>'"` consistently.
