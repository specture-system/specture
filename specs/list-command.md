---
number: 5
status: completed
author: Addison Emig
creation_date: 2026-02-05
approved_by: Addison Emig
approval_date: 2026-02-10
---

# List Command

Add a `specture list` command for querying specs with filtering and machine-readable output. This gives both agents and humans a quick way to see the state of all specs without reading individual files.

This is also a prerequisite for the [frontmatter numbering spec](/specs/frontmatter-numbering.md) — once spec numbers move out of filenames, the list command becomes the primary way to see spec numbers.

## Design Decisions

### Building on `status` internals

The [status command spec](/specs/status-command.md) defines parsing logic for extracting spec metadata: name, number, status (with inference), current task, task section, and complete/incomplete task lists. The `list` command should reuse this same parsing infrastructure, applying it across all specs rather than just the current one.

### Filtering

The command should support filtering by:

- `--status <value>` — filter by spec status (comma-separated for multiple: `draft,in-progress`)

No filters shows all specs.

### Output formats

- `--format text` (default) — human-readable table with columns: Number, Status, Progress (e.g., `3/7`), Name. Sorted by ascending spec number
- `--format json` — structured JSON array with full spec metadata (everything from the `status` command, per spec)

The JSON format is the primary interface for agents. It should include all fields defined in the status command spec: name, number, status, current task, current task section, complete tasks, and incomplete tasks.

### No task listing in text output

The text output is intentionally a compact overview — one row per spec. For task details, use `specture status --spec N`. The JSON format still includes full task information for programmatic use.

## Task List

### Core Implementation

- [x] Write tests for list command text and JSON output
- [x] Implement `specture list` command structure and aliases (`list`, `ls`)
- [x] Use `spec.ParseAll` from `internal/spec` ([spec 003](/specs/status-command.md)) to load and parse all specs
- [x] Implement text output with columns: Number, Status, Progress (e.g., `3/7`), Name — sorted by ascending spec number
- [x] Implement JSON output with full `SpecInfo` metadata per spec (name, number, status, current task, current task section, complete tasks, incomplete tasks)
- [x] Add `--format` flag (`text`, `json`)

### Filtering

- [x] Write tests for filtering (single status, multiple statuses, no matches)
- [x] Implement `--status` filter (single value) — uses resolved status from `SpecInfo`
- [x] Implement `--status` filter with comma-separated multiple values

### Documentation

- [x] Add usage examples to `specture list --help`
- [x] Include `list` in `specture help` workflow overview
- [x] Update the specture skill with `list` command usage
