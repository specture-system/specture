---
status: draft
author: Shelley
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

`specture setup` should handle migration of existing projects:

- Scan for files matching the `NNN-slug.md` pattern
- Extract the number from the filename
- Add `number` field to frontmatter if not already present
- Rename the file to remove the numeric prefix
- Report changes for user confirmation before applying

### Backward compatibility

During a transition period, the CLI should accept both formats:

- Files with numeric prefix and no `number` in frontmatter — number extracted from filename
- Files with `number` in frontmatter and no numeric prefix — number from frontmatter
- Files with both — frontmatter takes precedence, `specture validate` warns about the redundancy

### Cross-references

Specs that link to other specs by filename (e.g., `[spec 003](/specs/003-status-command.md)`) will break after migration. `specture setup` should update these links as part of the migration. Going forward, specs should reference other specs by number in prose (e.g., "see spec 3") since filenames are no longer stable identifiers for numbering.

## Task List

### Core Changes

- [ ] Add `number` as an optional frontmatter field
- [ ] Update spec parsing to read number from frontmatter, falling back to filename
- [ ] Update `specture new` to auto-assign next available number in frontmatter
- [ ] Update `specture new` to generate slug-only filenames (no numeric prefix)
- [ ] Update `specture validate` to detect duplicate numbers across specs
- [ ] Update `specture validate` to warn when number exists in both filename and frontmatter

### Migration

- [ ] Implement migration logic in `specture setup`: scan, extract, rename, update frontmatter
- [ ] Update cross-reference links in spec files during migration
- [ ] Add `--dry-run` support for migration preview
- [ ] Add user confirmation before applying migration changes

### Documentation

- [ ] Update `specs/README.md` template to reflect new file naming convention
- [ ] Update `specture help` to describe numbering in frontmatter
- [ ] Update spec template to include `number` field in frontmatter
