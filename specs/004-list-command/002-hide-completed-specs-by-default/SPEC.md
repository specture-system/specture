---
status: in-progress
author: Addison Emig
creation_date: 2026-06-18
---

# Hide Completed Specs By Default

Filter out completed specs from `list` output by default, since they're noise for daily work. Users can see completed specs by passing `--status completed`.

Only applies when no `--status` flag is given — if the user explicitly filters by status, respect it literally.

## Design Decisions

- **Reuse `--status completed` to opt in.** No new flag needed — the existing `--status` filter already lets users view completed specs explicitly.
- **`--status all` shows all statuses.** Useful as a blanket override to see everything, including completed specs.
- **Only applies in the default (no `--status`) case.** When `--status` is passed, the filter is used as-is without hiding completed.
