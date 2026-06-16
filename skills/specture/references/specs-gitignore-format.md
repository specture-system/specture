# specs/.gitignore Format

Use `specs/.gitignore` to keep scratch notes and implementation artifacts out of git while preserving durable specs and agent handoff plans.

Recommended content:

```gitignore
*
!*/
!**/SPEC.md
!**/PLAN.md
!README.md
```

## Why

- `*` ignores arbitrary scratch files under `specs/` by default.
- `!*/` allows Git to traverse nested spec directories.
- `!**/SPEC.md` tracks durable design records.
- `!**/PLAN.md` tracks agent execution handoffs.
- `!README.md` keeps local specs documentation visible.
