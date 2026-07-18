---
status: completed
author: Addison Emig
creation_date: 2026-07-18
approved_by: Addison Emig
approval_date: 2026-07-18
assignee: Addison Emig
---

# Assignee Support

Some projects already use an `assignee` frontmatter field to record who owns a spec, but Specture does not currently expose that metadata when querying the spec tree. This makes it impossible to use the CLI to find the specs assigned to a particular person.

## Goals

- Treat `assignee` as supported optional spec metadata.
- Allow specs to be filtered by assignee from the CLI.
- Display assignments in text list output without adding an empty column when assignees are not in use.

## Design Decisions

### Assignee matching

- Chosen: Let `--assignee` accept one or more comma-separated names and match a spec when its assignee case-insensitively equals any complete trimmed value.
  - This mirrors the existing `--status` filter, tolerates capitalization and surrounding whitespace, and supports querying several assignees without partial-name false matches.
- Considered: Match partial assignee names.
  - Partial matching could return unintended assignees.
- Considered: Accept only one assignee per invocation.
  - A single-value filter would be less useful and inconsistent with `--status`.

### Text list output

- Chosen: Add an `ASSIGNEE` column to `specture list` and its `ls` alias only when at least one spec in the final displayed result set includes an assignee.
  - This makes ownership visible while avoiding an empty column when assignees are not in use. Specs excluded by status, parent, depth, or assignee filters do not affect whether the column appears.
- Considered: Always include the `ASSIGNEE` column.
  - An always-present column would add noise to projects that do not use assignees.

### JSON output

- Chosen: Always include an `assignee` string in each entry from `specture list --format json`, using an empty string for an unassigned spec.
  - A consistent field gives agents and other consumers a stable schema.
- Considered: Omit `assignee` from entries for unassigned specs.
  - A conditional field would require consumers to handle two output shapes.

### Unassigned filtering

- Chosen: Do not add a reserved assignee value or a separate flag for filtering unassigned specs.
  - Filtering specifically for unassigned specs is outside this spec's scope.
