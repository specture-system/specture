---
status: approved
author: Addison Emig
creation_date: 2026-07-15
approved_by: Addison Emig
approval_date: 2026-07-15
---

# Default Depth to All

Change the default value of the `-d`/`--depth` flag from `1` to `all`, so bare `specture list` shows the entire spec tree instead of only the top level.

This supersedes the default established in [Add Depth Flag](specs/004-list-command/001-add-depth-flag/SPEC.md). That spec defaulted to `1` to avoid overwhelming the user, but [Hide Completed Specs By Default](specs/004-list-command/002-hide-completed-specs-by-default/SPEC.md) already filters out most of the noise. Showing the full tree by default is much more convenient — both for humans scanning the spec landscape and for agents that need full context.

## Design Decisions

- **Default `--depth` to `all`.** Full tree by default, for both humans and agents. The old concern about overwhelming output is mitigated by completed specs being hidden by default (4.2).
- **Supersedes the `--parent` exception from 4.1.** 4.1 special-cased `--parent` to default `--depth` to `all`. Since `all` is now the unconditional default, that exception is no longer needed.
- **`--depth 1` remains available** as the escape hatch for users who want the old compact, single-level view.
