---
status: completed
author: Addison Emig
creation_date: 2026-06-16
---

# Agent-Native Redesign

Specture is a bit bloated. It has too many features and tries to do too many things.

We should:

- Make it a lightweight CLI that integrates well with coding agents through a single robust skill.
- Remove extra code/features that we don't use often, because they are maintenance debt at this point every time we want to make any changes.

The CLI should:

- Focus on querying and validating `SPEC.md` and `PLAN.md` files
- Include a small helper for deterministically creating new specs

The agent skill should:

- Carry the workflow knowledge: how to bootstrap the spec tree, migrate older layouts, choose between specs and plans, and use `specture validate` as the verification step after file edits.

## Goals

- Keep Specture small enough that each command is easy to understand and maintain.
- Treat the Specture skill as the primary workflow layer for coding agents.
- Keep the CLI focused on querying, validation, and deterministic spec creation.
- Support existing project issue numbering systems by allowing explicit spec references.
- Remove command behavior that duplicates what coding agents can already do safely by editing files.

## Design Decisions

### Lightweight CLI with a single robust skill

- Chosen: Make the skill the workflow authority and the CLI the querying/validation helper.
  - Modern coding agents are already good at reading instructions, editing files, and applying migrations.
  - Specture remains useful as the reliable way to list, filter, inspect, and validate the spec tree.
  - The CLI should not carry expensive workflow automation that agents can perform from clear skill instructions.
  - This keeps Specture useful across different agent ecosystems instead of coupling it to one orchestration model.
  - The direction builds on [Workflow Assistance](specs/003-workflow-assistance/SPEC.md), but narrows the CLI responsibility further.
- Chosen: The skill should explain both `SPEC.md` and `PLAN.md`.
  - `SPEC.md` is the durable design record.
  - `SPEC.md` is more often human-written and deliberately planned out before implementation.
  - `PLAN.md` is the disposable execution handoff for a coding agent.
  - `PLAN.md` is more likely to be written by an LLM based on discussion with a human.
  - Keeping these separate avoids turning specs into stale implementation checklists.
  - This aligns with [Rework Spec Organization](specs/008-spec-hierarchy/SPEC.md), which moved specs away from granular task progress.
- Chosen: Organize the skill as a small entrypoint with task-specific reference files.
  - The skill should follow the [Agent Skills specification](https://agentskills.io/specification) so it works with existing skill installers and validators.
  - `SKILL.md` should stay concise and route agents to focused files under `references/` for planning, implementation, validation, migration, and spec/plan format guidance.
  - File references from `SKILL.md` should use relative paths from the skill root and stay one level deep to avoid nested reference chains.
  - Long-form examples, templates, and rarely needed procedures should live outside `SKILL.md` so agents only load the context needed for the current task.
  - The distributable skill should live at `skills/specture/` in this repository so skill installers can discover it as a standard skill package.
  - The repo-local `.agents/skills/specture/` copy should be removed so there is only one source of truth for the Specture skill.
  - This keeps one installable Specture skill for maintenance and distribution while avoiding an oversized default prompt.
- Considered: Keep growing CLI workflow commands.
  - More automation looks useful initially, but every workflow rule becomes long-term maintenance debt.
  - Agent behavior and conventions are moving quickly enough that hard-coding workflow orchestration into the CLI is the wrong abstraction.

### Remove `setup`

- Chosen: Remove the `setup` command and all setup-owned migration/install logic.
  - `setup` maintains expensive code for bootstrapping `specs/`, generating `specs/README.md`, installing skill files, and migrating older layouts.
  - Those operations are plain file edits that an agent can perform from skill instructions.
  - `specture validate` should be the verification boundary after the agent performs a migration.
- Chosen: Move migration guidance into the skill.
  - The skill should explain how to create a missing `specs/` tree.
  - The skill should explain how to migrate flat spec files into `SPEC.md` directories.
  - The skill should explain how to normalize cross-spec links to repo-root-relative `SPEC.md` paths.
  - The skill should explain that the agent must run `specture validate` after migrations.
- Considered: Keep `setup` only as a thin bootstrap command.
  - Even a thin command still needs tests, templates, overwrite behavior, and compatibility decisions.
  - Removing the command entirely creates a clearer boundary: agents edit files, Specture validates them.

### Simplify `new`

- Chosen: Make `specture new` non-interactive and file-only.
  - `--title` is required.
  - `--parent` optionally selects the parent spec scope.
  - `--spec` / `-s` optionally selects the new spec reference instead of auto-allocating the next number.
  - `--plan` creates `PLAN.md` instead of the default `SPEC.md`.
  - If `--spec` is omitted, Specture auto-allocates the next number in the selected scope.
  - The command creates only one file from the matching standard template.
- Chosen: Default `new` to `SPEC.md`, with explicit opt-in for `PLAN.md`.
  - `specture new --title "Feature"` creates `SPEC.md`.
  - `specture new --title "Feature" --plan` creates `PLAN.md`.
  - `PLAN.md` creation should use the same directory numbering, `--spec` reference, and parent-scope rules as `SPEC.md` creation.
  - A directory may contain both `SPEC.md` and `PLAN.md`.
  - Creating `PLAN.md` in an existing spec directory is valid when the user wants to add an execution handoff to a durable spec.
  - Creating `SPEC.md` in an existing plan directory is valid when the user wants to add a durable design record next to an existing plan.
  - The command should still fail if the target file already exists.
  - `PLAN.md` only needs spec frontmatter when it stands alone without `SPEC.md`; if both files exist, only `SPEC.md` is queried and validated.
- Chosen: Remove interactive behavior from `new`.
  - No title prompt.
  - No confirmation prompt.
  - No editor launch.
  - No stdin body support.
  - No dry-run mode.
- Chosen: Remove branch behavior from `new`.
  - No branch creation.
  - No `--no-branch` flag.
  - No clean-worktree requirement.
  - No branch cleanup logic.
  - Branch policy belongs to the project or agent workflow, not the spec file organizer.
- Chosen: Support explicit refs with `--spec` / `-s`.
  - Projects often have existing issue or ticket numbers that are useful to preserve in spec refs.
  - Reusing `--spec` keeps the CLI vocabulary consistent with commands that select existing spec references.
  - For `new`, `--spec` means the spec reference to create, not an existing spec to resolve.
  - `specture new --title "Login Timeout" --spec 123` creates `specs/123-login-timeout/SPEC.md`.
  - `specture new --title "Backend Contract" --spec 123.4` creates a child spec with full ref `123.4`.
  - Dotted `--spec` values include the full parent path and should resolve all parent specs before creating the child.
  - `--parent` remains useful when auto-allocating the next child ref under an existing parent.
  - `--spec` and `--parent` are mutually exclusive; if `--spec` is provided, the command should reject `--parent`.
- Considered: Keep stdin body support for generated specs.
  - Agents can edit the created file directly after generation.
  - Removing body input keeps the command smaller and avoids frontmatter/body merging behavior.
- Considered: Keep dry-run mode.
  - The command has a small, obvious effect and uses safe file creation.
  - Removing dry-run reduces flags and test surface.
