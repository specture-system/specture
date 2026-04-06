---
number: 8
status: completed
author: Addison Emig
creation_date: 2026-03-27
approved_by: Bennett Moore
approval_date: 2026-03-30
---

# Rework Spec Organization

This will be a large change to improve how our specs are organized. The goals and design decisions are based on an hour-long meeting we had about our experience with Specture and the different pain points we have experienced in the existing design.

## Goals

- Users should be allowed to nest specs to any number of levels.
  - The current flat organization system is tedious to work with.
- Specs should not contain step-by-step granular implementation progress.
  - If you try to preplan a giant task list to complete a spec, it will quickly become outdated.
  - Agents often forget to check off items in the task list as they go.
- Spec numbers should be immutable long term references.
  - They should no longer have any connection to priority or conflict resolution between specs.
- Update the skill to match the new design.
- Update the `specs/README.md` template to match the new design.
- Update the `help` subcommand output to match the new design.
- Update the validator to match the new design.

## Design Decisions

- Specs should be directories, with `SPEC.md` files.
  - Sub-specs are added as subdirectories within a spec directory, along with their own `SPEC.md` files.
  - This allows better modeling of large features, similar to GitLab epics.
  - A parent spec's status is managed manually — there is no automatic rollup from child statuses.
  - The full reference number of a spec will include all the numbers of its ancestors, separated by `.`.
    - For example, Spec 1.4.3.
    - This follows the typical pattern used in engineering design documents.
    - In code, `SpecInfo.Number` remains an `int` (the local/per-level number). A new `FullRef string` field (e.g., `"1.4.3"`) is computed during parsing by walking ancestors.
  - Sub-spec numbers are scoped per parent, not global.
    - `FindNextSpecNumber` allocates from the scope of the target parent (using `max(existing) + 1`, or `0` if no specs exist in that scope).
- Spec numbers should be included as a prefix for the directory names to allow easy scanning when using tools like `ls`.
  - Left padded `0` are optional.
  - For example, `specs/0-MVP/1-backend/13-auth/SPEC.md` and `specs/000-mvp/001-backend/013-auth/SPEC.md` are both valid.
  - The `number:` field stays in frontmatter; the directory prefix is for human scanning convenience only.
- `specture new` accepts a `--parent` flag with a full dotted reference (e.g., `--parent 1.4`) to create sub-specs.
  - Without `--parent`, specs are created at the top level.
- `specture list` shows only top-level specs by default.
  - Use `--parent 1.4` to list children of a specific spec.
- All commands that accept a `--spec` flag support dotted references (e.g., `1.4.3`).
  - This applies to `status`, `validate`, `list` (via `--parent`), `implement`, and `rename`.
  - `ResolvePath` is updated to resolve dotted references by walking the directory tree.
- `specture rename` updates the directory slug (e.g., `3-status-command/` → `3-new-name/`) and rewrites all repo-root-relative cross-spec links that reference the old path.
- The `specs/README.md` template should be simplified.
  - The directory tree is self-documenting; the README only needs a brief description and a link to the Specture repo.
- Validator should no longer check the `## Task List` section.
  - We won't error if it still exists (from old specs), but it should not be required in the future.
- `specture validate` (no args) validates the entire spec tree recursively. `specture validate --spec 1.4.3` validates only that single `SPEC.md` without recursing into children.
- Validator should no longer throw error `links: spec links must use the referenced spec title, not generic labels like 'spec 12' or 'spec #12'`
  - We now have immutable spec numbers, and they should be more permanent than spec titles.
- Cross-spec markdown links should use repo-root-relative file paths (e.g., `[Status command](specs/3-status-command/SPEC.md)`).
  - The migration automatically rewrites existing relative links to the new repo-root-relative paths.
- The `setup`/`update` subcommand should automatically reorganize the `specs` directory for the new directory and naming scheme.
  - Spec directories only track `SPEC.md` and `README.md` — no other assets for now.
    - Add `specs/.gitignore` for each project:
      ```
      *
      !**/SPEC.md
      !README.md
      ```
