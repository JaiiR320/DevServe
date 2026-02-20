# AGENTS.md - Agentic Coding Guidelines for devserve

## Project Overview
A Go CLI tool that serves local JavaScript development servers across Tailscale networks.

## Build/Lint/Test Commands

```bash
# Build the application
go build -o devserve

# Run the application
go run main.go

# Run all tests
go test ./...

# Run a single test
go test -run TestFunctionName ./path/to/package

# Run tests with verbose output
go test -v ./...

# Format code
go fmt ./...

# Vet code for issues
go vet ./...

# Tidy dependencies
go mod tidy

# Download dependencies
go mod download
```

## Code Style Guidelines

### Imports
- Group imports: standard library first, then blank line, then third-party packages
- Use explicit imports (no dot imports)
- Internal project imports use module path: `devserve/internal`

### Formatting
- Use `go fmt` for automatic formatting
- Use tabs for indentation (Go standard)
- Line length: aim for 100 characters, hard limit at 120
- No trailing whitespace

### Naming Conventions
- **Packages**: lowercase, single word, no underscores (e.g., `internal`, `cmd`)
- **Exported**: PascalCase (e.g., `Serve`, `DetectPackageManager`)
- **Unexported**: camelCase (e.g., `getOutputs`, `savePid`)
- **Variables**: descriptive names (e.g., `portStr`, `doneChan`)
- **Constants**: PascalCase for exported, camelCase for unexported
- **Interfaces**: noun ending in "-er" (e.g., `Reader`, `Writer`)
- **Acronyms**: all caps (e.g., `PID`, `URL`, `HTTP`)

### Types
- Prefer explicit types over `interface{}`
- Use structs for related data
- Return errors as last value: `(result, error)`

### Error Handling
- Always check errors explicitly
- Wrap errors with context: `fmt.Errorf("failed to X: %w", err)`
- Return errors rather than logging in library code
- Use `errors.New()` for static error messages
- Never ignore errors (use `_` only when truly safe)

### Functions
- Keep functions small and focused
- Early returns for error cases
- Document exported functions with comments starting with function name
- Limit parameters (consider structs for 3+ related params)

### Project Structure
```
.
├── cmd/           # Cobra CLI commands
│   ├── root.go    # Root command setup
│   ├── serve.go   # Serve subcommand
│   └── detect.go  # Detect subcommand
├── internal/      # Internal packages
│   ├── serve.go   # Serve logic
│   └── detect.go  # Package manager detection
├── main.go        # Entry point
├── go.mod         # Module definition
└── go.sum         # Dependency checksums
```

### Comments
- All exported items must have doc comments
- Comments start with the item name
- Use `//` for single-line, `/* */` for multi-line license headers only

### Testing
- Test files: `*_test.go`
- Test functions: `TestFunctionName(t *testing.T)`
- Table-driven tests preferred
- Use `t.Parallel()` for independent tests

### Dependencies
- Minimize external dependencies
- Current: `github.com/spf13/cobra` for CLI
- Go version: 1.25.5
