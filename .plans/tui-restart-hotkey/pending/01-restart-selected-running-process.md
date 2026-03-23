# Slice 1: Restart Selected Running Process

## Dependencies

None.

## Description

Pressing `r` on the selected row restarts a running process, whether it is saved or unsaved. The TUI reloads afterward, keeps the selection on the same process when possible, shows the action in the help bar, and reports success or failure in the status line.

## Expected Behaviors Addressed

- When I hover a running saved process and press `r`, it restarts and the status line confirms success or failure.
- When I hover a running unsaved process and press `r`, it also restarts using its current live command and directory.
- The help bar shows `r` as an available action.
- After the action, the list refreshes and keeps the selection on the same item when possible.

## Acceptance Criteria

- [ ] Pressing `r` on a running saved process stops it and starts it again.
- [ ] Pressing `r` on a running unsaved process stops it and starts it again using the selected row's current fields.
- [ ] The help text includes `r` with restart wording.
- [ ] The list reloads after restart and preserves the selected process when possible.
- [ ] The status line shows a success or failure message that matches existing TUI conventions.

## QA

1. Start the TUI with at least one running saved process visible in the list.
2. Hover that process and press `r`.
3. Confirm the process restarts and the status line reports the result.
4. Confirm the same process remains selected after the list reloads.
5. Repeat with a running unsaved process.
6. Confirm the help bar advertises the `r` hotkey.

---

*Appended after execution.*

## Completion

- Built TUI restart handling for the selected running row. Pressing `r` now stops the selected process, starts it again from the row's current fields, reloads the list, preserves selection by process name when the row still exists, updates the status line with restart success or failure, and shows `r restart` in the help bar.
- Kept restart as a separate row action in the main Bubble Tea key handler and reused the existing stop/start client paths. Running unsaved rows restart from their live command, directory, port, and name because the TUI passes the selected `listItem` back into the start path.
- Tightened the reload flow so row actions reload against the selected process name instead of only the previous cursor index. That keeps the same process focused after reorderings and still falls back to the previous cursor position if the process disappears.
- Added TUI unit coverage for running saved restart, running unsaved restart, restart failure messaging, `r` key dispatch, and help text.
- Deviation: this slice leaves `r` on stopped rows as a non-running error (`cannot restart '<name>': process is not running`). Slice 2 should replace that branch for saved stopped rows with a start action while keeping the running-row restart path unchanged.
