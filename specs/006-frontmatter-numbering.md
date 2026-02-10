---
status: approved
author: Addison Emig
creation_date: 2026-02-05
approved_by: Addison Emig
approval_date: 2026-02-10
---

# Frontmatter Numbering

Spec numbers are currently encoded in filenames (e.g., `003-status-command.md`). This causes problems:

- **Numbering collisions** — two people creating specs independently can pick the same number (this already happened with two `003-` specs in this repo)
- **Renaming files to reorder** is disruptive — breaks links, git history, and cross-references
- **The CLI parses filenames** to extract numbers, coupling file naming conventions to tool behavior

Move spec numbers into frontmatter and use plain slug filenames.

## Design Decisions

### Number in frontmatter, slug in filename

- Chosen: `number` field in YAML frontmatter, filename is just the slug
  - Filenames become: `status-command.md`, `workflow-assistance.md`
  - Frontmatter becomes: `number: 3`
  - `specture new` auto-assigns the next available number
  - `specture validate` detects duplicate numbers
  - `specture list` is the primary way to see spec numbers (per [spec 005](/specs/005-list-command.md))
  - Reordering precedence is a frontmatter edit, not a file rename
- Considered: Keep numbers in filenames
  - Simple and visible in file explorers
  - But causes the collision and renaming problems described above
- Considered: Remove numbers entirely, use creation date for precedence
  - Eliminates collisions
  - But loses the explicit precedence ordering that the Specture System relies on

### Migration of existing specs

`specture setup` should handle adding `number` to existing specs:

- Scan for files matching the `NNN-slug.md` pattern
- Extract the number from the filename
- Add `number` field to frontmatter if not already present
- Do **not** rename files — existing filenames stay as-is to preserve links, git history, and cross-references
- Report changes for user confirmation before applying (uses existing `--dry-run` and `--yes` flags)

### Number is a required field

`number` is a required frontmatter field — a non-negative integer (starting from 0). `specture validate` fails if any spec is missing it or has an invalid value. This keeps the system simple — one source of truth, no fallback logic, no ambiguity.

The path for existing projects is: run `specture setup`, which adds `number` to all specs that are missing it (extracted from the `NNN-` filename prefix). After that, validation passes.

### Mixed filename formats are fine

After migration, old files keep their `NNN-slug.md` filenames and new files use `slug.md`. Both naming patterns are valid — the CLI does not care about filenames, only frontmatter. Users can clean up old filenames at their own pace using `specture rename`.

### Rename command

`specture rename` updates a spec's filename and all markdown links that reference it across the specs directory.

```bash
specture rename --spec 3 --slug status-command
```

This would rename `003-status-command.md` to `status-command.md` and update any `[text](/specs/003-status-command.md)` links in other specs to `[text](/specs/status-command.md)`.

- `--dry-run` previews changes without applying them
- If `--slug` is omitted, the command strips the numeric prefix from the current filename
- The command updates all markdown links in `specs/` that reference the old filename

### Auto-assignment uses max+1

`specture new` assigns the next number as max(existing numbers) + 1. Gaps in numbering are allowed and not backfilled. This avoids confusion about which numbers are "available."

### Number/filename mismatch

If a file has a `NNN-` prefix and a `number` field that disagree, `specture validate` warns about the inconsistency. The frontmatter `number` is always authoritative.

## Task List

### Core Changes

- [ ] Tests for `number` parsing (present, missing, invalid values)
- [ ] Add `number` field to spec parsing, read exclusively from frontmatter
- [ ] Tests for all new validate rules (missing number, duplicates, number/filename mismatch)
- [ ] `specture validate` rejects specs missing `number`
- [ ] `specture validate` detects duplicate numbers across specs
- [ ] `specture validate` warns on number/filename mismatch
- [ ] Tests for new `specture new` behavior (auto-assign max+1, slug-only filename)
- [ ] `specture new` assigns max+1 number in frontmatter
- [ ] `specture new` generates slug-only filenames

### Rename

- [ ] Tests for rename command (file rename, link updates, --slug, --dry-run)
- [ ] `specture rename` renames file and updates all markdown links in specs directory
- [ ] `--slug` flag sets target filename; default strips numeric prefix
- [ ] `--dry-run` previews changes without modifying files

### Migration

- [ ] Tests for migration (adds number, skips existing, dry-run)
- [ ] `specture setup` adds `number` to frontmatter of `NNN-slug.md` files
- [ ] Migration respects existing `--dry-run` and `--yes` flags

### Documentation

- [ ] Update `specs/README.md` template to reflect new file naming convention
- [ ] Update `specture help` to describe numbering in frontmatter
- [ ] Update spec template to include `number` field in frontmatter
- [ ] Update `.skills/specture/SKILL.md` to document `number` frontmatter field and new filename conventions
