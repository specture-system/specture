---
status: draft
author: Addison Emig
creation_date: 2026-02-05
---

# List Command

Add a `specture list` command for querying specs with filtering and machine-readable output. This gives both agents and humans a quick way to see the state of all specs without reading individual files.

This is also a prerequisite for the [frontmatter numbering spec](/specs/006-frontmatter-numbering.md) — once spec numbers move out of filenames, the list command becomes the primary way to see spec numbers.

## Design Decisions

### Building on `status` internals

The [status command spec](/specs/003-status-command.md) defines parsing logic for extracting spec metadata: name, number, status (with inference), current task, task section, and complete/incomplete task lists. The `list` command should reuse this same parsing infrastructure, applying it across all specs rather than just the current one.

### Filtering

The command should support filtering by:

- `--status <value>` — filter by spec status (comma-separated for multiple: `draft,in-progress`)
- `--spec <number>` — show a single spec by number

Filters are composable. No filters shows all specs.

### Output formats

- `--format text` (default) — human-readable table with columns: Number, Status, Progress (e.g., `3/7`), Name. Sorted by ascending spec number
- `--format json` — structured JSON array with full spec metadata (everything from the `status` command, per spec)

The JSON format is the primary interface for agents. It should include all fields defined in the status command spec: name, number, status, current task, current task section, complete tasks, and incomplete tasks.

### Task listing

By default, the text output shows a compact overview (one row per spec). Additional flags expose task details:

- `--tasks` — include task lists in output
- `--incomplete` — only show incomplete tasks (implies `--tasks`)
- `--complete` — only show complete tasks (implies `--tasks`)

Task flags show top-level tasks only (consistent with the status command's treatment of indented tasks). The JSON format always includes full task information regardless of these flags.

## Task List

### Core Implementation

- [ ] Implement `specture list` command structure and aliases
- [ ] Reuse spec parsing infrastructure from `status` command
- [ ] Implement text output with columns: Number, Status, Progress, Name
- [ ] Implement JSON output with full spec metadata
- [ ] Add `--format` flag (`text`, `json`)

### Filtering

- [ ] Implement `--status` filter (single value)
- [ ] Implement `--status` filter with comma-separated multiple values
- [ ] Implement `--spec` filter by number

### Task Display

- [ ] Implement `--tasks` flag to include task lists in text output
- [ ] Implement `--incomplete` flag (only incomplete tasks, implies `--tasks`)
- [ ] Implement `--complete` flag (only complete tasks, implies `--tasks`)

### Documentation

- [ ] Add usage examples to `specture list --help`
- [ ] Include `list` in `specture help` workflow overview
