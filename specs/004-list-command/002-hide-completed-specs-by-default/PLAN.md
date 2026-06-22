# Hide Completed Specs By Default Plan

Implement [Hide Completed Specs By Default](specs/004-list-command/002-hide-completed-specs-by-default/SPEC.md) in one focused change.

## Pull Request Plan

### PR 1: Make list hide completed specs by default

- Add command tests in `cmd/list_test.go` proving default text output excludes specs with `status: completed`.
- Add JSON coverage proving default `--format json` output also excludes completed specs.
- Preserve explicit status filtering: `--status completed` should return completed specs, and `--status draft,in-progress` should behave literally without applying any implicit completed filter.
- Add `--status all` support as the blanket override that returns every status, including completed specs.
- Update list help text in `cmd/list.go` so the default behavior and `--status all` override are visible in `specture list --help`.
- Keep the filtering orchestration in `cmd/list.go`; do not move CLI-specific default behavior into `internal/spec`.

## Implementation Notes

- Apply the implicit filter only when the `--status` flag is omitted.
- Treat `all` as a special status-filter value in the list command, not as a new spec status.
- Make `all` work consistently for text and JSON because both output formats share the same filtered spec slice.
- Revisit existing tests named around "all specs" or sorted default output; update their expectations so "all" means explicit `--status all` rather than default output.
- Run `just test` after implementation.
- Run `just run validate --spec 4.2` after any plan or spec edits.
