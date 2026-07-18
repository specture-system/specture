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

Optional fields include `author`, `assignee`, `creation_date`, `approved_by`, and `approval_date`.

## Body

Start with a single H1 title. Use the structure that matches the spec's role in the hierarchy.

### Parent Specs

A spec with child specs is a concise index for that feature area. It contains a simple description, goals, and a `## Child Specs` section with repo-root-relative links to its direct children.

Parent structure:

```markdown
# Parent Feature

Describe the overall feature area simply.

## Goals

- Goal one
- Goal two

## Child Specs

- [First Child](specs/012-parent/000-first-child/SPEC.md)
- [Second Child](specs/012-parent/001-second-child/SPEC.md)
```

Do not put design decisions in a parent spec. Record each decision in the child spec that owns the relevant design. Update the child index when adding, renaming, moving, or removing a direct child.

### Leaf Specs

A spec without children contains the detailed design. Include only goals, requirements, and design decisions the user explicitly discussed or confirmed. Do not infer decisions merely to complete the recommended structure.

Recommended leaf structure:

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
- Keep parent specs limited to a simple description, goals, and linked child-spec index; keep design decisions in child specs.
- Do not add goals, requirements, or design decisions that the user has not explicitly discussed or confirmed.
- Keep task checklists and execution notes in `PLAN.md`, not `SPEC.md`.
- Use repo-root-relative markdown links for cross-spec references, such as `[Status command](specs/002-status-command/SPEC.md)`.
- Run `specture validate` after edits.
