---
number: 7
status: in-progress
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
- Additionally: The commit-sized orchestration unit is the top-level checkbox in a task-list section
  - Nested checkboxes and nested bullets at any depth are part of the parent task's acceptance criteria, not separate orchestration units
  - Worker and reviewer prompts must include the full nested subtree for the parent task across all nesting levels
  - Accepting a parent task means marking its full nested subtree complete in the same spec update, including descendants at every depth
- Considered: Allow arbitrarily large tasks to span many commits
  - Weakens the task-to-commit linkage that the command is designed to enforce

### Sectioned task-list requirement

- Chosen: Require every top-level task-list checkbox to appear under a `###` section
  - Keeps section planning and branch sequencing deterministic
  - Prevents ambiguous unsectioned tasks from bypassing section-based orchestration
  - Allows nested checkbox trees while still assigning each top-level task to one section
- Considered: Allow unsectioned top-level tasks in `## Task List`
  - Makes section-based planning and validation less reliable

### Dry-run scope

- Chosen: `implement --dry-run` computes and prints the execution plan, then exits before making any changes
  - Preserves the safety expectation that dry-run mode performs no git mutations, agent invocations, spec edits, commits, or pushes
  - Still lets users validate backend selection and remaining work before execution
- Considered: Simulate the full orchestration loop while only suppressing writes
  - Adds complexity without improving the core planning preview use case

### Final cleanup pass

- Chosen: After all sections are complete, run one final cleanup review and one cleanup worker pass, then create a final refactor commit
  - Provides one bounded opportunity to simplify or polish the completed implementation
  - Keeps cleanup deterministic by avoiding retries and follow-up review loops
  - Focuses cleanup on unnecessary abstraction, clear `AGENTS.md` guideline violations, and low-risk maintainability improvements
- Considered: Reuse the same retry-and-review loop as section execution
  - Turns final cleanup into another open-ended gate instead of a bounded polish pass

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

- [x] Add failing tests for deterministic section branch naming, clean-worktree checks, and fail-closed rerun behavior
- [x] Implement section branch creation, clean-worktree enforcement, and rerun validation
- [x] Add failing tests for worker invocation and task-level retry behavior when review finds critical issues
- [x] Implement worker-agent invocation that passes the current task, section context, and spec path, and instructs workers to avoid editing the spec or creating commits
- [x] Implement task-level review-agent invocation and rerun the worker up to 3 total passes only when review finds critical issues

### Spec Updates and Section Delivery

- [x] Add failing tests for `in-progress` transition, task checkbox updates, and deterministic task commits
- [x] Implement spec status updates and task completion commits so each accepted task updates the spec in the same commit as its implementation
- [x] Add failing tests for section-level review, single-retry behavior, strict push gating, and push failure handling
- [x] Implement section-level review with exactly 1 revision retry on critical issues, then push the completed section branch and stop immediately if that push fails

### Task Structure and Validation

- [x] Update the `implement` command to support nested checkboxes in the task list
  - [x] Treat each top-level checkbox as one implementation, review, and commit unit
  - [x] Include nested checkboxes and nested bullets at every depth in the worker and reviewer context for the parent top-level task
  - [x] Mark the full nested subtree complete when the parent top-level task is accepted, including descendants at every depth
- [x] Update the `validate` command to require every top-level task-list checkbox to appear under a `###` section

### CLI Polish

- [x] Improve the help message for the `implement` command
  - [x] Mention that it is an agent orchestrator
  - [x] Include example usage
- [x] Add `--dry-run` support to `implement` so it prints the execution plan and exits before making changes
- [x] Rename the new implement prompt templates to use `task-` and `section-` prefixes, and update related code
- [x] Improve task execution progress output so worker and reviewer passes are visible while `implement` runs
- [ ] Show reviewer feedback output for each pass so multi-pass retries are diagnosable from CLI logs

### Final Cleanup

- [ ] Add a final "clean up" stage after all sections are completed
  - [ ] Run one final cleanup review for refactor opportunities in the completed work
  - [ ] Focus the cleanup review on unnecessary abstraction, clear `AGENTS.md` guideline violations, and low-risk maintainability improvements
  - [ ] Run one cleanup worker pass to implement the recommended refactors
  - [ ] Create a final refactor commit for the cleanup pass
  - [ ] Do not retry or re-review after the cleanup pass

### Completion

- [ ] Add failing tests for final `completed` status handling
- [ ] Implement the final completion update when all remaining tasks are done
