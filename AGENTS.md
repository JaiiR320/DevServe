# DevServe Architecture

## Learnings

### Error handling conventions
Wrap errors with `fmt.Errorf("failed to <verb> <noun>: %w", err)`. Always single-quote user-provided names in errors (e.g. `"process '%s' not found"`). Use `"failed to <verb>"` prefix for error log messages. Keep log-then-return pattern in handlers. Validation errors use `"missing or invalid '<field>'"` consistently.
