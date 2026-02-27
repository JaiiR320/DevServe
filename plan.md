# DevServe Test Suite — Product Requirements Document

## Overview

DevServe currently has **zero test coverage**. This PRD defines a phased plan to introduce a comprehensive test suite across all packages, working from the lowest-level pure functions up to integration-level daemon tests. Each feature represents a logically grouped set of tests that can be committed independently.

### Architecture Reference

```
CLI (cmd/)  ──Unix Socket──>  Daemon (daemon/)
   │                              │
   └── uses internal/             └── uses internal/
       (client, protocol,             (protocol, process,
        style, config)                 port, config, files,
                                       logger)
```

### Workflow

- Each **Feature** below is a commit-sized unit of work.
- Within each feature, **Tasks** break the work into individual test functions or small refactors.
- When a feature (or a large task with many changes) is done, **suggest a git commit** with a short summary. Do not commit automatically.
- After a commit is accepted and created, mark the feature and its tasks as complete.

---

## Feature 1: Protocol Encoding/Decoding Tests

> **File:** `internal/protocol_test.go`
> **Source under test:** `internal/protocol.go`
> **Commit message:** `test: add protocol request/response encoding tests`

The protocol layer is the foundation of all daemon communication. These are pure functions operating over `net.Conn` with JSON encoding — no side effects, no filesystem, no network.

### Tasks

- [x] **1.1** Test `OkResponse` — returns `Response{OK: true, Data: data}`
- [x] **1.2** Test `ErrResponse` — returns `Response{OK: false, Error: err.Error()}`
- [x] **1.3** Test `SendRequest` / `ReadRequest` round-trip — use `net.Pipe()`, send a `Request` with action and args, read it back, assert equality
- [x] **1.4** Test `SendResponse` / `ReadResponse` round-trip — use `net.Pipe()`, send a `Response`, read it back, assert equality
- [x] **1.5** Test `ReadRequest` with malformed JSON — write garbage bytes to a pipe, assert error is returned with `"failed to decode request"` message
- [x] **1.6** Test `ReadResponse` with malformed JSON — write garbage bytes to a pipe, assert error is returned with `"failed to decode response"` message
- [x] **1.7** Test `Request` with nil/empty `Args` — ensure `omitempty` behavior works (args omitted from JSON when nil)
- [x] **1.8** Test `Response` with empty `Data`/`Error` — ensure `omitempty` behavior works

---

## Feature 2: File Utility Tests

> **File:** `internal/files_test.go`
> **Source under test:** `internal/files.go`
> **Commit message:** `test: add file utility tests for LastNLines`

`LastNLines` is a self-contained file reading utility. Tests use `os.CreateTemp` for isolation.

### Tasks

- [x] **2.1** Test `LastNLines` with more lines than N — write 10 lines, request 5, assert last 5 returned
- [x] **2.2** Test `LastNLines` with fewer lines than N — write 3 lines, request 10, assert all 3 returned
- [x] **2.3** Test `LastNLines` with exactly N lines — write 5 lines, request 5, assert all 5 returned
- [x] **2.4** Test `LastNLines` with empty file — create empty file, assert `nil, nil` returned
- [x] **2.5** Test `LastNLines` with nonexistent file — pass bogus path, assert `nil, nil` returned (no error)
- [x] **2.6** Test `LastNLines` with trailing newlines — write content with extra trailing `\n`, assert no empty trailing line
- [x] **2.7** Test `LastNLines` with single line, no trailing newline — assert single-element slice returned

---

## Feature 3: Port Utility Tests

> **File:** `internal/port_test.go`
> **Source under test:** `internal/port.go`
> **Commit message:** `test: add port checking and waiting utility tests`

Port utilities require real TCP listeners. Tests start/stop `net.Listen("tcp", ...)` to simulate occupied and free ports.

### Tasks

- [x] **3.1** Test `CheckPortInUse` when port is free — pick an unused port, assert no error
- [x] **3.2** Test `CheckPortInUse` when port is occupied — start a TCP listener, assert error contains `"already in use"`
- [x] **3.3** Test `WaitForPort` when port is immediately available — start listener before calling, assert returns quickly with no error
- [x] **3.4** Test `WaitForPort` when port becomes available after a delay — start listener in a goroutine after 200ms, assert returns successfully
- [x] **3.5** Test `WaitForPort` timeout — never start a listener, use a short timeout (e.g. 300ms), assert error contains `"not ready after"`

---

## Feature 4: CLI Styling Tests

> **File:** `internal/style_test.go`
> **Source under test:** `internal/style.go`
> **Commit message:** `test: add CLI styling and table rendering tests`

Formatting functions are pure (string in, string out). Tests validate structure rather than exact ANSI escape sequences since terminal rendering varies.

### Tasks

- [x] **4.1** Test `Success` — assert output contains `"✓"` and the message string
- [x] **4.2** Test `Error` — assert output contains `"✗"` and the message string
- [x] **4.3** Test `Info` — assert output contains `"•"` and the message string
- [x] **4.4** Test `RenderTable` with valid JSON array — `[{"name":"app","port":3000}]`, assert output contains `"NAME"`, `"PORT"`, `"app"`, `"3000"`
- [x] **4.5** Test `RenderTable` with multiple entries — assert all entries appear in output
- [x] **4.6** Test `RenderTable` with empty array `[]` — assert output contains `"No active processes"`
- [x] **4.7** Test `RenderTable` with invalid JSON — assert raw input is returned as fallback
- [x] **4.8** Test `HelpTemplate` — assert returns non-empty string containing `"Usage:"` and `"Commands:"`

---

## Feature 5: Process Lifecycle Tests + Tailscale Refactor

> **Files:** `internal/process.go` (refactor), `internal/process_test.go` (new)
> **Source under test:** `internal/process.go`
> **Commit message:** `refactor: make tailscale call injectable; add process lifecycle tests`

`Process.Start()` and `Process.Stop()` both shell out to `tailscale serve`. To test process lifecycle without requiring Tailscale, we refactor the Tailscale calls into an injectable function (package-level variable or struct field). Tests replace it with a no-op.

### Tasks

- [ ] **5.1** Refactor: extract Tailscale `serve --https ... --bg` call in `Start()` into an injectable function (e.g. `TailscaleServe func(port int) error` field on `Process`, or a package-level `var RunTailscaleServe`)
- [ ] **5.2** Refactor: extract Tailscale `serve --https ... off` call in `Stop()` into an injectable function (e.g. `TailscaleStop func(port int) error` field on `Process`, or pair with above)
- [ ] **5.3** Test `CreateProcess` — use `t.TempDir()`, assert `.devserve/` directory created, assert `out.log` and `err.log` files exist
- [ ] **5.4** Test `CreateProcess` with invalid directory — assert error returned with `"failed to create log directory"` message
- [ ] **5.5** Test `Start` with a simple command — use a command like `python3 -m http.server <port>` or `nc -l <port>`, inject no-op Tailscale func, assert process is running and port becomes ready
- [ ] **5.6** Test `Stop` after `Start` — start a process, stop it, assert process exited and logs are closed
- [ ] **5.7** Test `Stop` idempotency — call `Stop()` twice, assert second call returns `"already stopped"` error
- [ ] **5.8** Test `Stop` before `Start` — call `Stop()` on a fresh process, assert `"has not been started"` error

---

## Feature 6: Daemon Handler Tests

> **Files:** `daemon/handlers.go` (minor export), `daemon/handlers_test.go` (new)
> **Source under test:** `daemon/handlers.go`
> **Commit message:** `test: add daemon handler tests with test helpers`

Handlers operate on the package-level `processes` map and `mu` mutex. We need to either export these or add test-helper functions so tests can set up and tear down state. Since handlers call `internal.CheckPortInUse` and `internal.CreateProcess`/`Process.Start`, some tests will need real ports or mocks.

### Tasks

- [ ] **6.1** Add test helpers: export or create helper functions to initialize/reset the `processes` map and `mu` for tests (e.g. `ResetProcesses()`, `SetProcess(name, *Process)`, `GetProcesses()`)
- [ ] **6.2** Test `handlePing` — assert returns `OkResponse` with data `"pong"`
- [ ] **6.3** Test `handleServe` with missing `name` arg — assert error response with `"missing or invalid 'name'"`
- [ ] **6.4** Test `handleServe` with missing `port` arg — assert error response
- [ ] **6.5** Test `handleServe` with missing `command` arg — assert error response
- [ ] **6.6** Test `handleServe` with duplicate name — pre-populate processes map, assert error `"already in use"`
- [ ] **6.7** Test `handleServe` port type handling — test with `float64` (JSON number) and `string` port values
- [ ] **6.8** Test `handleServe` with invalid port type — assert error `"invalid port type"`
- [ ] **6.9** Test `handleStop` with missing name — assert error response
- [ ] **6.10** Test `handleStop` with nonexistent process — assert error `"not found"`
- [ ] **6.11** Test `handleList` with empty map — assert returns `OkResponse` with `"[]"` JSON
- [ ] **6.12** Test `handleList` with populated map — pre-populate, assert JSON contains entries
- [ ] **6.13** Test `handleLogs` with missing name — assert error response
- [ ] **6.14** Test `handleLogs` with nonexistent process — assert error `"not found"`
- [ ] **6.15** Test `handleLogs` with valid process — create process with log files containing known content, assert output contains stdout/stderr sections

---

## Feature 7: Client & Daemon Integration Tests

> **Files:** `internal/client_test.go`, `daemon/daemon_test.go`
> **Source under test:** `internal/client.go`, `daemon/daemon.go`
> **Commit message:** `test: add client and daemon integration tests`

Integration tests exercise the full Unix socket communication path. These are heavier, may need custom socket paths (to avoid conflicting with a real daemon), and test the `handleConn` dispatch logic end-to-end.

### Tasks

- [ ] **7.1** Test `Send` returns `ErrDaemonNotRunning` when no socket exists — call `Send()` with a non-existent socket path, assert error wraps `ErrDaemonNotRunning`
- [ ] **7.2** Test `handleConn` dispatches `ping` action — set up a Unix socket listener, send a ping request through it, assert `"pong"` response
- [ ] **7.3** Test `handleConn` with unknown action — send a request with action `"bogus"`, assert error response contains `"unknown action"`
- [ ] **7.4** Test `handleConn` with malformed request — send garbage bytes, assert error response
- [ ] **7.5** Test `handleConn` dispatches `shutdown` — send shutdown request, assert response indicates success and stop channel is signaled
- [ ] **7.6** Test `stopAllProcesses` with empty map — assert returns nil (no failures)
- [ ] **7.7** Test full daemon foreground lifecycle (optional/advanced) — start `startForeground()` in a goroutine with a test socket path, send ping, send shutdown, assert clean exit

---

## Summary

| Feature | Scope | Includes Refactor? |
|---------|-------|--------------------|
| 1. Protocol Tests | `internal/protocol.go` | No |
| 2. File Utility Tests | `internal/files.go` | No |
| 3. Port Utility Tests | `internal/port.go` | No |
| 4. Styling Tests | `internal/style.go` | No |
| 5. Process Tests | `internal/process.go` | Yes — injectable Tailscale calls |
| 6. Handler Tests | `daemon/handlers.go` | Yes — export test helpers for process map |
| 7. Integration Tests | `internal/client.go`, `daemon/daemon.go` | No |

**Total test functions:** ~45
**Estimated files created:** 7 test files + minor edits to 2 source files
