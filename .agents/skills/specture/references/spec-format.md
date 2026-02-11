# Spec File Format Reference

Detailed format specification for Specture spec files. Loaded on demand when creating or editing specs.

## File Location and Naming

Specs live in the `specs/` directory with a 3-digit numeric prefix and kebab-case name:

```
specs/000-mvp.md
specs/001-add-authentication-system.md
specs/013-refactor-database-layer.md
specs/314-redesign-api-endpoints.md
```

The numeric prefix determines precedence (higher number = higher precedence). Numbers should generally increment with each new spec.

## Complete Example

```markdown
---
status: draft
author: Your Name
creation_date: 2025-12-18
---

# Feature or Change Name

Description as paragraphs and/or bulleted list.

## Design Decisions

### Decision Title

- Chosen: Option A
  - Pro 1
  - Pro 2
- Considered: Option B
  - Pro 1
  - Con 1

## Task List

### Foundation

- [ ] Task 1
- [ ] Task 2
- [ ] Task 3

### Core Implementation

- [ ] Task 1
- [ ] Task 2

### Polish and Documentation

- [ ] Task 1
- [ ] Task 2
```

## Frontmatter

YAML frontmatter between `---` delimiters at the top of the file.

### Required Fields

| Field    | Values                                                         |
| -------- | -------------------------------------------------------------- |
| `status` | `draft`, `approved`, `in-progress`, `completed`, or `rejected` |

### Optional Fields

| Field           | Format       | Description                          |
| --------------- | ------------ | ------------------------------------ |
| `author`        | Free text    | Person(s) who proposed/wrote the spec |
| `creation_date` | `YYYY-MM-DD` | Date the spec was created            |
| `approved_by`   | Free text    | Person(s) who approved the spec      |
| `approval_date` | `YYYY-MM-DD` | Date the spec was approved           |

### Status Values

- **`draft`** — Being written and refined. May go through multiple iterations.
- **`approved`** — Team has agreed on the design; ready for implementation.
- **`in-progress`** — Implementation underway. Tasks are being checked off.
- **`completed`** — All tasks in the task list are done.
- **`rejected`** — Reviewed and rejected. Document **why** if merging a rejected spec.

## Required Sections

### Title (H1)

A clear, descriptive `# Heading` summarizing what is being proposed. Must be the first heading in the document.

### Description

Overview of the proposed change, immediately after the title. Can be paragraphs or bulleted lists. Consider including:

- What is being proposed
- Why it's needed
- What problem it solves
- High-level approach

For large descriptions, use additional `##` sections (e.g., `## Ideas`, `## Goals`, `## Benefits`).

### Task List (`## Task List`)

Implementation tasks as markdown checklists. Group related tasks under `###` subsections.

```markdown
## Task List

### Phase 1

- [ ] Task description
- [ ] Another task

### Phase 2

- [ ] More tasks
```

**Task list best practices:**

- Make tasks specific and actionable
- Order by dependencies (prerequisites first)
- Group related tasks into `###` sections
- Include testing, documentation, and deployment tasks
- Keep individual tasks reasonably sized (one commit each)
- Avoid implementation-level detail — describe *what*, not *how*

**During implementation:**

- Check off tasks by changing `- [ ]` to `- [x]`
- Add new tasks or sections as implementation reveals needs
- Remove or update tasks that turn out to be unnecessary

## Optional Sections

### Design Decisions

Document major design choices with options considered and their trade-offs:

```markdown
## Design Decisions

### Choice Title

- Chosen: Option A
  - Reason it was selected
  - Another advantage
- Considered: Option B
  - Advantage of this option
  - Why it was not selected
```

Include as many decision points as needed. No obligation for small or trivial specs.

## Precedence Rules

1. Higher-numbered specs take precedence over lower-numbered specs when they conflict on any point.
2. Completed specs are historical records — do **not** retroactively update them.
3. Exception: fix typos, documentation errors, or factual inaccuracies in completed specs.
4. When a new spec supersedes part of an older spec, the new spec's rules apply.

## Scope Guidelines

**Specs are for planned changes**: features, refactors, redesigns, tooling improvements.

**Not for bugs**: If software doesn't match what specs describe, use the issue tracker.

**When to split specs**: Changes are independent, can be deployed separately, solve different problems.

**When to combine**: Changes share design decisions or have task dependencies.

## Validation

Run `specture validate` to check specs against these format rules. The validator checks:

- Valid YAML frontmatter with required `status` field
- Status is one of the allowed values
- Description section is present (content after H1 title)
- Task list section is present (`## Task List`)
