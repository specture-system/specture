# SPEC.md Format

`SPEC.md` is the durable design record for a planned change. It should explain why the change exists and what design choices were made, not track implementation progress.

## Location

Specs live under `specs/` as directories containing `SPEC.md` files:

```text
specs/011-agent-native-redesign/SPEC.md
specs/011-agent-native-redesign/001-child-work/SPEC.md
```

Spec refs are derived from directory names. Do not store a spec number in frontmatter.

## Frontmatter

Required:

```yaml
---
status: draft
---
```

Valid statuses are `draft`, `approved`, `in-progress`, `completed`, and `rejected`.

Optional fields include `author`, `creation_date`, `approved_by`, and `approval_date`.

## Body

Start with a single H1 title, followed by a description and any useful sections.

Recommended structure:

```markdown
# Feature Name

Describe the problem, motivation, and high-level approach.

## Goals

- Goal one
- Goal two

## Design Decisions

### Decision Title

- Chosen: Selected option
  - Why it was selected
- Considered: Alternative option
  - Why it was not selected
```

## Rules

- Do not number markdown headings.
- Keep task checklists and execution notes in `PLAN.md`, not `SPEC.md`.
- Use repo-root-relative markdown links for cross-spec references, such as `[Status command](specs/002-status-command/SPEC.md)`.
- Run `specture validate` after edits.
