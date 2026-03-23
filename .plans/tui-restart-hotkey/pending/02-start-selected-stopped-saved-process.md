# Slice 2: Start Selected Stopped Saved Process

## Dependencies

Slice 1: Restart Selected Running Process.

## Description

Pressing `r` on a saved process that is not currently running starts it using its saved configuration. The TUI reloads afterward, keeps the selection on the same process when possible, and reports success or failure in the status line.

## Expected Behaviors Addressed

- When I hover a stopped saved process and press `r`, it starts the process.
- After the action, the list refreshes and keeps the selection on the same item when possible.

## Acceptance Criteria

- [ ] Pressing `r` on a saved process that is not running starts it with its saved command, directory, port, and name.
- [ ] The list reloads after the start attempt and preserves the selected process when possible.
- [ ] The status line shows a success or failure message that matches existing TUI conventions.
- [ ] Restart behavior for running rows from slice 1 remains unchanged.

## QA

1. Start the TUI with a saved process that is not currently running.
2. Hover that process and press `r`.
3. Confirm the process starts and appears as running after the list reloads.
4. Confirm the status line reports the result.
5. Re-test `r` on a running process to confirm restart behavior still works.

---

*Appended after execution.*

## Completion

- Updated the TUI `r` row action so stopped saved rows now use the existing start path instead of returning a non-running restart error. The action starts the selected process from its saved `name`, `port`, `command`, and `directory`, then reloads the list and preserves selection by process name through the existing reload flow.
- Kept slice 1 behavior intact for running rows. Running saved and unsaved rows still follow the separate restart path that stops first, then starts from the selected row's current fields, and still report restart-specific success and failure messages.
- Reused the existing TUI status-message conventions instead of inventing restart-specific wording for stopped rows. Successful starts report `process '<name>' started`, and stopped saved start failures report `failed to start '<name>': ...`.
- Added TUI unit coverage for the new stopped-saved branch and its failure path, while keeping the existing running-row restart tests in place to guard against regressions.
- Verified the slice with `go test ./tui` and `go test ./...`.
- Deviation: none for this slice. If a later slice touches restart behavior again, keep the split semantics: `r` restarts running rows but starts stopped saved rows, while stopped unsaved rows remain unsupported because they are not reachable in the current list model.
