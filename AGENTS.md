# AGENTS.md - Agentic Coding Guidelines for devserve

## Project Overview

A Go CLI tool that starts local JavaScript development servers and exposes them
across Tailscale networks. Uses Cobra for CLI scaffolding. The repo includes a
minimal JS fixture (`package.json`, `index.html`, `bun.lock`) so the tool can
self-test by detecting "bun" as the package manager.

## Build / Lint / Test Commands

```bash
# Build
go build -o devserve

# Run directly
go run main.go

# Run all tests
go test ./...

# Run a single test by name
go test -run TestFunctionName ./path/to/package

# Run tests with verbose output
go test -v ./...

# Format all code
go fmt ./...

# Static analysis
go vet ./...

# Tidy module dependencies
go mod tidy
```

There is currently no Makefile, no golangci-lint config, and no test files in the
repository. When adding tests, place them in `*_test.go` alongside the code they
exercise.

## Project Structure

```
.
├── main.go              # Entry point — calls cmd.Execute()
├── cmd/
│   ├── root.go          # Root Cobra command ("devserve")
│   ├── start.go         # start subcommand — starts dev server + Tailscale
│   ├── stop.go          # stop subcommand — kills process + disables Tailscale
│   ├── list.go          # list subcommand — shows running processes
│   └── detect.go        # detect subcommand — prints detected package manager
├── internal/
│   ├── config.go        # Lookup maps (LockToPM, PMToCommand)
│   ├── detect.go        # DetectPackageManager — checks lockfiles
│   ├── start.go         # Start() — orchestrates PM detection, process, Tailscale
│   ├── processes.go     # Process struct, Start/Wait/Stop, port polling
│   ├── registry.go      # JSON-based process registry (save/get/list/remove)
│   ├── files.go         # FileManager — log files, .devserve directory
│   └── tailscale.go     # TailscaleManager — wraps `tailscale serve` CLI
├── go.mod               # Module: devserve, Go 1.25.5
├── go.sum
├── package.json         # Test fixture (bun dev script)
├── index.html           # Test fixture (hello-world page)
└── bun.lock             # Test fixture (triggers bun detection)
```

Runtime data is stored in `.devserve/` (gitignored). Process JSON files and log
files live there.

## Code Style

### Imports

Group in two blocks separated by a blank line:

1. Standard library **and** internal project imports (`devserve/internal`, `fmt`, `os`, …)
2. Third-party imports (`github.com/spf13/cobra`)

```go
import (
	"devserve/internal"
	"fmt"

	"github.com/spf13/cobra"
)
```

No dot imports. No aliased imports unless necessary to resolve a name collision.

### Formatting

- `go fmt` is the authority — use tabs, not spaces.
- Aim for ≤100 characters per line; hard limit at 120.

### Naming

| Kind              | Convention       | Examples                                     |
|-------------------|------------------|----------------------------------------------|
| Packages          | lowercase, one word | `cmd`, `internal`                          |
| Exported symbols  | PascalCase       | `Start`, `CreateProcess`, `FileManager`      |
| Unexported symbols| camelCase        | `saveProcess`, `generateProcess`, `basePath` |
| Acronyms          | ALL CAPS         | `PID`, `PM`, `URL`                           |
| Receivers         | short (1–2 chars)| `p` (Process), `fm` (FileManager), `tm` (TailscaleManager) |
| Variables         | short, descriptive | `portStr`, `sigChan`, `doneChan`           |
| Constructors      | `NewXxx` / `CreateXxx` | `NewLocalFM()`, `CreateProcess()`      |

### Types and Structs

- Prefer concrete types over `interface{}` / `any`.
- Return `(result, error)` — error is always the last return value.
- Use pointer receivers for methods that mutate or are large structs.
- JSON tags on struct fields follow `json:"fieldName"`. Exclude fields from
  serialization with `json:"-"`.

### Error Handling

- Always check errors; never silently discard them (use `_` only when
  intentionally ignoring, e.g. polling a port).
- In **library code** (`internal/`): return wrapped errors.
  ```go
  return fmt.Errorf("failed to start process: %w", err)
  ```
- In **command handlers** (`cmd/`): print the error and return.
  ```go
  fmt.Println(err)
  return
  ```
- Use `errors.New()` for static error strings.
- Keep error messages lowercase and without trailing punctuation, per Go
  convention.

### Comments

- Exported functions, types, and package-level variables should have a doc
  comment starting with the symbol name:
  ```go
  // Start detects the package manager, starts the dev server, and serves it over Tailscale.
  func Start(port int, bg bool) error {
  ```
- License headers use `/* */` block comments (Cobra generator boilerplate).
- Use `//` for all other comments.

### Functions

- Keep functions short and focused — prefer early returns for error cases.
- Limit positional parameters; use a struct when there are 3+ related params.
- Cobra command `init()` functions register flags and add the command to
  `rootCmd`.

## Testing Conventions

No tests exist yet. When adding them:

- Place test files next to the code: `internal/start_test.go`.
- Name test functions `TestXxx(t *testing.T)`.
- Prefer table-driven tests for multiple cases.
- Use `t.Parallel()` for tests with no shared state.
- Use `t.Helper()` in test helper functions.

## Dependencies

- **Single direct dependency:** `github.com/spf13/cobra` for CLI.
- Keep the dependency footprint minimal — prefer stdlib solutions.

## Git Conventions

- Commit messages: lowercase, imperative mood, short (no period).
  Examples: `add list command`, `fix stop command to disable tailscale serve`,
  `refactor process management with port checking and wait`.
