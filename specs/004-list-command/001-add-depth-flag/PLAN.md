# Add Depth Flag Plan

Implement [Add Depth Flag](specs/004-list-command/001-add-depth-flag/SPEC.md) in one focused change.

## Pull Request Plan

### PR 1: Add depth-aware list querying

- Add reusable depth-aware spec querying in `internal/spec`, keeping the existing direct-scope behavior available for callers that still need it.
- Add table-driven tests for top-level depth, parent-relative depth, `all`, and `0` as unlimited.
- Add `-d`/`--depth` to `specture list`.
- Parse depth in `cmd/list.go`, with default `1` unless `--parent` is set, where the default is `all`.
- Reject invalid depth values with a clear CLI error.
- Update `specture list --help` examples to show depth usage.
- Add command tests covering text and JSON output with depth filtering.

## Implementation Notes

- Keep reusable traversal and depth logic in `internal/spec`; keep `cmd/list.go` focused on CLI flag handling and output.
- Preserve existing `list` behavior when `--depth` is omitted and `--parent` is not set.
- Preserve existing `--parent` behavior except for the new default of unlimited depth.
- Run `just test` after implementation.
