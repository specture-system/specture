---
number: 3
status: completed
author: Addison Emig
creation_date: 2026-01-20
approved_by: Addison Emig
approval_date: 2026-02-06
---

# Status Command

Some common questions come up while using Specture to implement a spec:

- What's the current spec?
- What's next for this spec?
- How much progress have we made for this spec?

We should add a `status` command to allow quickly answering these questions without requiring the user to directly read any of the spec files.

## Design Decisions

### Output Format

By default, the output should be in human-readable plain text.

For easy processing by automated tooling we should also support `json` output.

We can use a `--format` flag with values `text` or `json`

### Output Contents

It is useful to include the following:

- Spec name (from top-level header)
- Spec number (from filename)
- Spec status
  - If not specified in frontmatter, we can automatically deduce status using the following simple algorithm:
    - Spec has no task list -> `draft`
    - Spec has no complete tasks -> `draft`
    - Spec has mixture of complete and incomplete tasks -> `in-progress`
    - Spec has only complete tasks -> `completed`
- Current task
  - The contents of first line beneath `## Task List` that starts with `- [ ]`
  - Do not include indented tasks
  - Empty string if no such line is found
- Current task section title
  - Determined by looking for first incomplete task in the task list, then moving up line-by-line to find a section header
  - Empty string if no section header is found between `## Task List` and the incomplete item
- Complete tasks
  - Parse the task list from the markdown and return every one that has been checked off
- Incomplete tasks
  - Parse the task list from the markdown and return every one that has not been checked off

In the future, we may more items to the status output. The current list should be a good start, we aren't trying to be comprehensive with everything that might end up being useful in the status command output in the long run. Future specs can suggest additions to the output.

### Current Spec

By default, this command will return the results for the "current spec".

The current spec is determined by sorting the specs by ascending spec number, then selecting the first that has status `in-progress`.

Status `in-progress` can be inferred even for specs without explicit `status` value in their front matter using the algorithm mentioned above.

A `--spec` flag can be used to get the overall status of any particular spec by number, no matter if it is `in-progress` or not.

## Task List

### Shared Spec Package (`internal/spec`)

Build a shared `internal/spec` package that consolidates all spec parsing, discovery, and querying. This replaces the ad-hoc parsing scattered across `internal/validate/parser.go` (goldmark-based parsing) and `cmd/validate.go` (file discovery). The `validate` command, `status` command, and future `list` command ([spec 005](/specs/list-command.md)) will all consume this package.

- [x] Create `SpecInfo` struct with fields: Path, Name, Number, Status (resolved), CurrentTask, CurrentTaskSection, CompleteTasks, IncompleteTasks
- [x] Create `Task` struct with fields: Text, Complete, Section
- [x] Move `findAllSpecs` and `resolveSpecPath` from `cmd/validate.go` into `internal/spec` as `FindAll` and `ResolvePath`
- [x] Implement `Parse(path) (*SpecInfo, error)` — wraps goldmark parsing from `internal/validate/parser.go`, extends it with task extraction and status inference
- [x] Implement `ParseAll(specsDir) ([]*SpecInfo, error)` — parses all specs in directory, sorted by ascending number
- [x] Implement task list parser: extract top-level complete and incomplete tasks from the `## Task List` section (skip indented sub-tasks)
- [x] Implement current task detection: first `- [ ]` line under `## Task List`, excluding indented lines
- [x] Implement current task section detection: scan upward from current task to find nearest `### ` heading under `## Task List`
- [x] Implement status inference algorithm: no task list → use frontmatter or `draft`; no complete tasks → `draft`; mixed → `in-progress`; all complete → `completed`; explicit frontmatter status always overrides inference
- [x] Implement `FindCurrent(specs []*SpecInfo) *SpecInfo` — returns first spec with resolved status `in-progress`, sorted by ascending number
- [x] Write tests for task parsing (complete, incomplete, mixed, empty, indented sub-tasks)
- [x] Write tests for current task and current task section detection (happy path, no tasks, no section header)
- [x] Write tests for status inference (all combinations of frontmatter status × task states)
- [x] Write tests for `FindAll`, `ResolvePath`, `ParseAll`, `FindCurrent`
- [x] Refactor `cmd/validate.go` to use `internal/spec` for file discovery (`FindAll`, `ResolvePath`)
- [x] Refactor `internal/validate` to reuse shared parsing where possible (validator keeps its own validation logic, but delegates parsing to `internal/spec`)

### Command Implementation

- [x] Add `status` command with alias `s` and wire into root command
- [x] Implement default behavior: find current in-progress spec via `FindCurrent`, display its status
- [x] Add `--spec` flag to target a specific spec by number (reuses `ResolvePath`)
- [x] Implement `--format text` output (human-readable, default)
- [x] Implement `--format json` output (structured JSON, for tooling and agents)
- [x] Handle edge cases: no specs directory, no in-progress spec, spec not found, empty task list
- [x] Write tests for the status command (text output, JSON output, `--spec` flag, error cases)
