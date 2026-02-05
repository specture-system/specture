---
status: draft
author: Addison Emig
creation_date: 2026-02-05
---

# Workflow Assistance

The Specture System relies on agents reading and obeying instructions in `AGENTS.md` and `specs/README.md`. In practice, agents often forget to check off tasks in spec files and commit the spec update alongside their implementation changes.

Rather than relying on documentation compliance, we should make the CLI the primary interface for understanding and following the workflow:

- A `done` command that makes checking off tasks trivially easy
- A pre-commit hook that reminds agents about spec updates at the moment they commit
- Simplified `AGENTS.md` prompt that points agents to the CLI instead of docs
- Workflow knowledge moved from `specs/README.md` into CLI help output

## Design Decisions

### `done` command with text matching

- Chosen: `specture done [substring]` — matches a task by text content
  - Task text is a stable identifier (unlike position or number)
  - Default (no argument) checks off the first unchecked task ("current task")
  - Substring match handles the common case without requiring exact text
  - After checking off the task, automatically runs `git add` on the spec file so it's staged for the next commit
- Considered: `specture done --task <number>`
  - Task numbers are fragile — tasks get added, removed, and reordered
  - Agent would need to count lines to determine the right number
- Considered: `specture complete` as the command name
  - `done` is shorter and more natural
  - `complete` could be confused with shell completion

### Pre-commit hook as a soft nudge

- Chosen: Non-blocking pre-commit hook (prints reminder, exits 0)
  - Commit always succeeds — no workflow disruption
  - Agents typically read `git commit` output, so the reminder is visible
  - No false-positive blocking on intermediate commits (refactoring, docs, etc.)
  - Only triggers when a spec has status `in-progress`
- Considered: Blocking pre-commit hook (exit 1 if spec not staged)
  - Too aggressive — many commits are legitimately unrelated to spec tasks
  - Would require escape hatches (`--no-verify`) that become habitual
  - Intermediate commits during a task would be blocked
- Considered: No hook, rely on `specture status` output for reminders
  - Only works if agents are already calling `specture status`
  - Misses the natural nudge point (commit time)

### Inferring in-progress status

The hook and `done` command should use the same status inference algorithm defined in the [status command spec](/specs/003-status-command.md):

- No task list → `draft`
- No complete tasks → `draft`
- Mix of complete and incomplete → `in-progress`
- All tasks complete → `completed`

This means specs don't need an explicit `status: in-progress` in frontmatter for the workflow tools to activate.

### CLI as the source of truth

Currently, agents must read `AGENTS.md` and `specs/README.md` to understand the Specture workflow. This is unreliable — agents may not read the docs, may misinterpret them, or the docs may be stale.

- Chosen: Move workflow knowledge into CLI help output, simplify docs to pointers
  - The `AGENTS.md` / `CLAUDE.md` prompt becomes: `This project uses the Specture System. Use \`specture help\` for more information.`
  - `specs/README.md` stays in every project but becomes minimal — a brief overview linking to the [Specture GitHub repo](https://github.com/specture-system/specture) and referencing the CLI for detailed usage
  - All detailed workflow information moves into `specture help` output
  - Less important or subcommand-specific information moves into `specture <command> --help` to avoid overwhelming agents with `specture help`
  - The CLI is always up-to-date (installed version = current instructions), eliminating doc drift
  - Agents only need to discover one thing (the `specture` command) instead of reading multiple files
- Considered: Keep detailed `AGENTS.md` prompt and comprehensive `specs/README.md`
  - Relies on agents reading and following docs
  - Prompt goes stale when Specture evolves
  - Different projects may have outdated versions of the prompt

### Current spec selection

The `done` command operates on the "current spec" by default — the first spec (by ascending spec number) with status `in-progress`. A `--spec` flag allows targeting a different spec.

## Task List

### `done` Command

- [ ] Implement `specture done` with no arguments (checks off first unchecked task in current spec)
- [ ] Implement substring matching for `specture done [text]`
- [ ] Handle ambiguous matches (multiple tasks match substring) with clear error message
- [ ] Handle no match with clear error message
- [ ] Run `git add` on the spec file after checking off the task
- [ ] Add `--spec` flag to target a specific spec instead of current
- [ ] Add `--dry-run` flag to preview changes without modifying files
- [ ] Print confirmation with task text, spec name, and "next task" hint

### Pre-commit Hook

- [ ] Implement hook logic: detect in-progress specs, check if spec file is staged
- [ ] Print soft reminder with spec name, current task, and `specture done` hint
- [ ] Always exit 0 (non-blocking)
- [ ] Silence the hook when no spec is in-progress
- [ ] Integrate with pre-commit framework (`.pre-commit-hooks.yaml`)

### Simplify Agent Prompt

- [ ] Reduce `agent-prompt.md` template to a minimal one-liner pointing to `specture help`
- [ ] Slim down `specs/README.md` template to a brief overview linking to the Specture GitHub repo and referencing the CLI
- [ ] Move important workflow information from `specs/README.md` template into `specture help` output
- [ ] Distribute subcommand-specific details into `specture <command> --help` messages
- [ ] Keep `specture help` focused and concise — avoid overwhelming agents
- [ ] Update `specture setup` to generate the simplified docs
