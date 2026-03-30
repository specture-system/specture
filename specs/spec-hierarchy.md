---
number: 8
status: draft
author: Addison Emig
creation_date: 2026-03-27
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
  - Sub-specs are added as subdirectories with a spec directory, along with their own `SPEC.md` files.
  - This allows better modeling of large features, similar to GitLab epics.
  - The full reference number of a spec will include all the numbers of its ancestors, separated by `.`.
    - For example, Spec 1.4.3.
    - This follows the typical pattern used in engineering design documents.
- Spec numbers should be included as a prefix for the directory names to allow easy scanning when using tools like `ls`.
  - Left padded `0` are optional
  - For example, `specs/0-MVP/1-backend/13-auth/SPEC.md` and `specs/000-mvp/001-backend/013-auth/SPEC.md` are both valid.
- Validator should no longer check the `## Task List` section.
  - We won't error if it still exists (from old specs), but it should not be required in the future.
- Validator should no longer throw error `links: spec links must use the referenced spec title, not generic labels like 'spec 12' or 'spec #12'`
  - We now have immutable spec numbers, and they should be more permanent than spec titles.
- The `setup`/`update` subcommand should automatically reorganize the `specs` directory for the new directory and naming scheme.
  - Also, add `specs/.gitignore` for each project:
    ```
    *
    !**/SPEC.md
    !README.md
    ```

## Task List

Leaving this as blank for now, to appease the current validator. I'll remove it before marking the spec complete.
