# Design Workflow

Use this workflow when creating a new spec or refining an existing `draft` spec.
When splitting an existing spec, use [split-workflow.md](split-workflow.md) instead.

## Steps

1. Create or select a branch for design work, using the project's branch naming conventions when they exist.
2. Use `specture new --title "Feature name"` when creating a new spec file.
3. Populate only the problem, goals, requirements, and decisions the user has explicitly discussed or confirmed. Do not infer missing design content or fill the template speculatively.
4. Identify unresolved design decisions and work through them interactively with the user. Prefer one meaningful decision at a time, explain the relevant trade-offs, and record the outcome only after the user confirms it.
5. Use the parent or leaf structure from [spec-format.md](spec-format.md), based on the spec's role in the hierarchy.
6. Keep task checklists and transient implementation notes out of `SPEC.md`.
7. Run `specture validate` after editing specs.
8. Commit and open a PR for review when the design is ready.

## Status Guidance

- `draft`: still being written or debated.
- `approved`: design is accepted and ready for implementation.
- `rejected`: design was considered and rejected; document why if it will be merged.

Do not mark a spec `in-progress` until implementation begins.
