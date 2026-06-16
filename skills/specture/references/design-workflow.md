# Design Workflow

Use this workflow when creating a new spec or refining an existing `draft` spec.

## Steps

1. Create or select a branch for design work, using the project's branch naming conventions when they exist.
2. Use `specture new --title "Feature name"` when creating a new spec file.
3. Write the durable design record in `SPEC.md`:
   - problem or motivation
   - goals
   - key design decisions
   - considered alternatives and trade-offs
4. Keep task checklists and transient implementation notes out of `SPEC.md`.
5. Run `specture validate` after editing specs.
6. Commit and open a PR for review when the design is ready.

## Status Guidance

- `draft`: still being written or debated.
- `approved`: design is accepted and ready for implementation.
- `rejected`: design was considered and rejected; document why if it will be merged.

Do not mark a spec `in-progress` until implementation begins.
