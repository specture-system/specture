---
status: draft
author: Addison Emig
creation_date: 2026-02-05
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

`number` is a required frontmatter field. `specture validate` fails if any spec is missing it. This keeps the system simple — one source of truth, no fallback logic, no ambiguity.

The path for existing projects is: run `specture setup`, which adds `number` to all specs that are missing it (extracted from the `NNN-` filename prefix). After that, validation passes.

## Task List

### Core Changes

- [ ] Add `number` as a required frontmatter field
- [ ] Update spec parsing to read number exclusively from frontmatter
- [ ] Update `specture validate` to require `number` in frontmatter
- [ ] Update `specture validate` to detect duplicate numbers across specs
- [ ] Update `specture new` to auto-assign next available number in frontmatter
- [ ] Update `specture new` to generate slug-only filenames (no numeric prefix)

### Migration

- [ ] Implement migration logic in `specture setup`: scan for `NNN-slug.md` files, extract number, add `number` field to frontmatter
- [ ] Use existing `--dry-run` and `--yes` flags for preview and confirmation

### Documentation

- [ ] Update `specs/README.md` template to reflect new file naming convention
- [ ] Update `specture help` to describe numbering in frontmatter
- [ ] Update spec template to include `number` field in frontmatter
