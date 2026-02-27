# DevServe Architecture

## Learnings

### CLI styling with lipgloss
Styled output uses `cli/style.go` helpers (`Success`, `Error`, `Info`, `RenderTable`). Uses `charmbracelet/lipgloss` with ANSI color codes (1=red, 2=green, 6=cyan). Keep styling minimal: colored icon prefix (`✓`, `✗`, `•`) + plain message text. Cobra errors are silenced (`SilenceErrors: true`) and handled in `Execute()` with `cli.Error()`.

### Error handling conventions
Wrap errors with `fmt.Errorf("failed to <verb> <noun>: %w", err)`. Always single-quote user-provided names in errors (e.g. `"process '%s' not found"`). Use `"failed to <verb>"` prefix for error log messages. Keep log-then-return pattern in handlers. Validation errors use `"missing or invalid '<field>'"` consistently.

### Package Structure
Root-level packages for clear separation of concerns:
- `cli/` - UI styling and logging utilities
- `daemon/` - Daemon server, protocol types, and client
- `process/` - Process lifecycle and port management  
- `tunnel/` - Tunnel interface (DI-ready) and Tailscale implementation
- `util/` - Configuration constants and file utilities
