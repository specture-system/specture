# Spec Guidelines

This document outlines the structure and best practices for writing spec files in Specture.

## Overview

Specs are living documents that describe planned changes to the system. They serve as a blueprint for implementation and a historical record of design decisions.

## Spec File Structure

Each spec file should be a markdown document in the `specs/` directory with the following sections:

### 1. Title

The spec should start with a clear, descriptive H1 heading that summarizes what is being proposed.

```markdown
# Feature Name or Change Description
```

### 2. Brief Description

A concise overview (2-4 paragraphs) that explains:
- What is being proposed
- Why it's needed
- What problem it solves
- High-level approach

### 3. Design Decisions

This section documents the design exploration process. For each major decision point:

**Format:**
```markdown
## Design Decisions

### Decision Point Name

Brief context about what needs to be decided.

#### Option 1: [Name]
**Pros:**
- Advantage 1
- Advantage 2

**Cons:**
- Disadvantage 1
- Disadvantage 2

#### Option 2: [Name]
**Pros:**
- Advantage 1
- Advantage 2

**Cons:**
- Disadvantage 1
- Disadvantage 2

**Selected Approach:** [Chosen option and rationale]
```

Include as many decision points as needed. This creates a valuable historical record of why certain choices were made.

### 4. Task List

A detailed breakdown of implementation tasks using markdown checklists. Split into logical sections if needed.

**Format:**
```markdown
## Task List

### Phase 1: Foundation
- [ ] Task 1
- [ ] Task 2
- [ ] Task 3

### Phase 2: Core Implementation
- [ ] Task 1
- [ ] Task 2

### Phase 3: Polish & Documentation
- [ ] Task 1
- [ ] Task 2
```

**Task List Best Practices:**
- Make tasks specific and actionable
- Order tasks logically (dependencies first)
- Group related tasks into sections
- Include testing, documentation, and deployment tasks
- Keep individual tasks reasonably sized (completable in one session)

## File Naming

Use descriptive, kebab-case filenames:
- `add-authentication-system.md`
- `refactor-database-layer.md`
- `redesign-api-endpoints.md`

## Workflow

1. **Create Spec**: Write the spec file with all required sections
2. **Submit PR**: Open a pull request adding the spec to `specs/`
3. **Discussion**: Team discusses and refines the spec in PR comments
4. **Approval**: Once approved, merge the spec PR
5. **Implementation**: Work through the task list, checking off items as completed
6. **Update**: Keep the spec updated as implementation reveals new details

## Optional Sections

Specs may include additional sections as needed:
- **Dependencies**: External libraries, services, or other specs
- **Security Considerations**: Security implications and mitigations
- **Performance Impact**: Expected performance characteristics
- **Migration Strategy**: For changes affecting existing systems
- **Testing Strategy**: Approach to testing the change
- **Rollback Plan**: How to revert if needed
- **Open Questions**: Unresolved items requiring further discussion

## Example Spec Template

```markdown
# [Feature/Change Name]

## Description

[2-4 paragraphs describing what, why, and how]

## Design Decisions

### [Decision Point 1]

Context for this decision.

#### Option A: [Name]
**Pros:**
- Pro 1
- Pro 2

**Cons:**
- Con 1
- Con 2

#### Option B: [Name]
**Pros:**
- Pro 1
- Pro 2

**Cons:**
- Con 1
- Con 2

**Selected Approach:** Option A - [rationale]

## Task List

### Phase 1: Setup
- [ ] Task 1
- [ ] Task 2

### Phase 2: Implementation
- [ ] Task 3
- [ ] Task 4

### Phase 3: Finalization
- [ ] Task 5
- [ ] Task 6
```

## Tips

- **Be clear, not clever**: Write for future readers who may not have context
- **Document alternatives**: Even rejected options are valuable to record
- **Update as you go**: Specs should evolve during implementation
- **Link to discussions**: Reference PR comments, issues, or other specs
- **Focus on "why"**: The code shows "how", the spec should explain "why"
