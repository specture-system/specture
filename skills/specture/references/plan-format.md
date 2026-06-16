# PLAN.md Format

`PLAN.md` is an execution handoff for coding agents. It can be more tactical and temporary than `SPEC.md`.

## Location

Place `PLAN.md` beside the relevant `SPEC.md` when the plan implements a durable spec:

```text
specs/011-agent-native-redesign/SPEC.md
specs/011-agent-native-redesign/PLAN.md
```

A standalone plan may exist without a sibling `SPEC.md` when the work does not yet have a durable design record. In that case, include enough frontmatter for Specture to query and validate it.

## Recommended Body

```markdown
# Feature Name Plan

Implement [Feature Name](specs/011-feature-name/SPEC.md) in small, reviewable chunks.

## Pull Request Plan

### PR 1: First reviewable slice

- Task one
- Task two

### PR 2: Follow-up slice

- Task one
- Task two

## Implementation Notes

- Constraints or sequencing details agents should preserve.
```

## Rules

- Keep plans actionable and easy to update.
- Link to the relevant `SPEC.md` when one exists.
- Do not duplicate design rationale that belongs in `SPEC.md`.
- It is acceptable to update or replace `PLAN.md` as execution details change.
