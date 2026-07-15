---
status: approved
author: Addison Emig
creation_date: 2026-07-15
approved_by: Addison Emig
approval_date: 2026-07-15
---

# Show Paths in List Output

Normalize the `Path` field in `SpecInfo` to be repo-root-relative (e.g., `specs/004-list-command/SPEC.md`) instead of an absolute filesystem path (e.g., `/home/addison/repos/open-source/specture/specs/004-list-command/SPEC.md`).

The `list` command already has a PATH column in its text output and a `path` field in JSON output, both sourced from `SpecInfo.Path`. Today the value is an absolute path, which is long, machine-specific, and not portable. A repo-root-relative path is shorter, consistent across checkouts, and directly usable in markdown links.

## Design Decisions

- **Normalize `SpecInfo.Path` to repo-root-relative.** The path stored on `SpecInfo` should be relative to the repository root (e.g., `specs/004-list-command/SPEC.md`), not an absolute filesystem path. This makes the value portable and consistent regardless of where the repo is checked out.
- **Applies to both text and JSON output.** Both the table PATH column and the JSON `path` field read from `SpecInfo.Path`, so normalizing the field fixes both formats with a single change.
