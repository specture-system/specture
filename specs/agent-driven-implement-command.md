---
number: 7
status: approved
author: Addison Emig
creation_date: 2026-03-12
approved_by: Addison Emig
approval_date: 2025-03-14
---

# Agent-Driven Implement Command

Add a new `implement` subcommand to the Specture CLI that orchestrates implementation of an `approved` or `in-progress` spec using agent CLIs.

Unlike a single handoff command, this workflow should break implementation into remaining task-list sections, create a branch per section, run worker and review agents in a deterministic loop, update the spec file itself, and push completed section branches automatically.

The command should prefer `opencode` when auto-detecting available agent CLIs, fall back to `codex`, and allow explicit override via `--agent`.

## Design Decisions

### Explicit spec selection with optional agent override

- Chosen: Require `--spec` and make `--agent` optional
  - Spec execution is high-impact and should target a specific spec explicitly
  - Agent backend can be discovered automatically without adding user friction
  - Auto-detection order is `opencode`, then `codex`
- Considered: Auto-pick a spec when `--spec` is omitted
  - Risky for a command that creates branches, commits, and pushes
  - Makes failures harder to reason about when multiple specs are available

### Multi-agent orchestration instead of a single implementation handoff

- Chosen: Run repeated worker and review agent invocations for each remaining section and task
  - Matches the desired workflow of task-by-task validation and deterministic commits
  - Keeps commits small and traceable to specific spec tasks
  - Allows a stricter review gate before each task and section is accepted
- Considered: One-shot agent execution for the full spec
  - Harder to verify progress incrementally
  - More likely to drift from the spec or produce oversized changes

### Same backend for worker and reviewer roles

- Chosen: Use the same selected backend for both worker and review runs
  - Keeps the command surface small in v1
  - Avoids separate backend configuration and compatibility logic
- Considered: Separate worker/reviewer agent flags
  - Adds complexity without a clear need for the initial implementation

### Section-based branch strategy

- Chosen: Create a branch for each remaining section, with later section branches based on the previously completed section branch
  - Preserves the exact workflow requested for staged section delivery
  - Allows each completed section to be pushed immediately while carrying earlier work forward
  - Reruns fail closed by default and only resume when the expected section branch and checked task state match unambiguously
- Considered: Branch every section from the original base branch
  - Would require replaying or merging prior section work into later branches
  - Makes sequential implementation more cumbersome

### Specture owns spec edits and commits

- Chosen: Worker agents must not edit the spec or commit; the orchestrator updates the spec and creates deterministic commits
  - Keeps workflow enforcement in one place
  - Ensures spec checkbox updates are always committed together with implementation changes
  - Makes commit messages and git mutations predictable
- Considered: Let workers edit the spec and commit directly
  - Harder to guarantee consistent spec/task bookkeeping
  - Makes the orchestration loop less deterministic

### Review retry policy

- Chosen: Allow up to 3 worker passes per task, and only 1 retry for section-level review
  - Gives tasks enough room to converge automatically
  - Prevents section-level review loops from becoming nit-picky or over-optimized
  - Review failure only means critical issues were found; minor concerns and nitpicks do not block progress
- Considered: Multiple section-level retries
  - Increases runtime and churn with diminishing returns

### Strict push gating between sections

- Chosen: Require each completed section branch to push successfully before starting the next section
  - Keeps local and remote branch progression aligned for PR review as implementation advances
  - Avoids later section branches depending on unpublished local history
- Considered: Continue to later sections after a push failure
  - Leaves remote review state out of sync with local section sequencing
  - Makes staged PR review less reliable

### Task sizing for deterministic commits

- Chosen: Specs intended for `implement` should define tasks as roughly one commit-sized unit of work
  - Matches the workflow expectation that each completed task produces a deterministic commit with the corresponding spec checkbox update
  - Keeps task review and retry loops focused on a single unit of progress
- Considered: Allow arbitrarily large tasks to span many commits
  - Weakens the task-to-commit linkage that the command is designed to enforce

### Spec state updates during implementation

- Chosen: Set `status: in-progress` when section execution begins if the spec was `approved`, then mark each task complete immediately after that task passes review
  - Reflects active work as soon as implementation starts
  - Preserves the Specture rule that spec updates ship with the implementation they describe
- Additionally: Set `status: completed` in the final successful commit when all remaining tasks are checked off
  - Keeps the spec lifecycle accurate at the end of orchestration
- Considered: Delay checkbox updates until section completion
  - Loses the per-task commit linkage required by the workflow

## Task List

### CLI and Planning

- [x] Add failing tests for `implement` command validation and allowed spec statuses
- [x] Add `specture implement --spec N [--agent opencode|codex]` to satisfy those tests
- [x] Add failing tests for loading a spec and enumerating remaining sections and tasks
- [x] Implement remaining-section and remaining-task planning
- [x] Add failing tests for backend auto-detection and `--agent` override
- [x] Implement backend selection with priority order `opencode`, then `codex`

### Branch and Task Execution

- [ ] Add failing tests for deterministic section branch naming, clean-worktree checks, and fail-closed rerun behavior
- [ ] Implement section branch creation, clean-worktree enforcement, and rerun validation
- [ ] Add failing tests for worker invocation and task-level retry behavior when review finds critical issues
- [ ] Implement worker-agent invocation that passes the current task, section context, and spec path, and instructs workers to avoid editing the spec or creating commits
- [ ] Implement task-level review-agent invocation and rerun the worker up to 3 total passes only when review finds critical issues

### Spec Updates and Section Delivery

- [ ] Add failing tests for `in-progress` transition, task checkbox updates, and deterministic task commits
- [ ] Implement spec status updates and task completion commits so each accepted task updates the spec in the same commit as its implementation
- [ ] Add failing tests for section-level review, single-retry behavior, strict push gating, and push failure handling
- [ ] Implement section-level review with exactly 1 revision retry on critical issues, then push the completed section branch and stop immediately if that push fails

### Completion

- [ ] Add failing tests for final `completed` status handling
- [ ] Implement the final completion update when all remaining tasks are done
