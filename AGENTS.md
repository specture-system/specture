# AGENTS.md

## Overview

Specture is a spec-driven software architecture system. It provides a lightweight, document-driven approach to project planning where specs are markdown files in the `specs/` directory.

**For detailed information about the Specture System, see `specs/README.md`.**

## Directory Structure

- **`specs/`**: Markdown spec files for planned changes (features, refactors, improvements)
- **`specs/README.md`**: Complete spec guidelines, workflow, and best practices
- **`AGENTS.md`**: This file, for agentic coding tools

## Core Concepts

### What Are Specs?

Specs are design documents describing planned changes: new features, major refactors, redesigns, and tooling improvements. They are inspired by PEP (Python Enhancement Proposal) and BIP (Bitcoin Improvement Proposal) systems.

- **NOT for bugs**: Use the issue tracker for bugs (cases where software doesn't match spec descriptions)
- **Scope**: Can range from large (traditional epic size) to small (minor UI improvements)
- **Coherent units**: Split specs for independent changes; combine for shared design decisions

### Spec Structure

All specs follow this format:

1. **YAML Frontmatter**: Metadata (status, author, dates)
2. **Title (H1)**: Clear, descriptive name
3. **Description**: Overview of what's being proposed, why, and the problem it solves
4. **Design Decisions** (optional): Rationale for major design choices
5. **Task List**: Detailed, actionable implementation tasks grouped into sections

### Spec Lifecycle

Each spec has a status field that tracks its state:

- **`draft`**: Being written and refined
- **`approved`**: Ready for implementation
- **`in-progress`**: Implementation underway
- **`completed`**: All tasks finished
- **`rejected`**: Reviewed but not approved

### File Naming

Use kebab-case filenames with numeric prefix:

- `000-mvp.md`
- `001-add-authentication-system.md`
- `013-refactor-database-layer.md`

### Precedence System

Higher-numbered specs take precedence over lower-numbered specs. If two specs conflict, the higher number wins. This avoids the need to retroactively update completed specs.

### Workflow Guidelines

1. **Check spec status** before starting work—only implement specs marked `approved` or `in-progress`
2. **Update task lists** as you complete tasks in a spec
3. **Don't update completed specs** (except for typos/factual corrections)—create a new spec if requirements change
4. **Document your work** in the task list; the spec becomes the historical record
5. **Refer to design decisions** in the spec when making implementation choices

See `specs/README.md` for detailed guidelines on spec workflow, best practices, and design decision documentation.

## Development Setup

Before starting work, ensure pre-commit hooks are installed:

```bash
pip install pre-commit
pre-commit install
```

This runs `just check` (formatting, linting, tests) before every commit.

## Build/Test Commands

Use `just` to run development tasks. See `justfile` for available recipes. Common commands:

- `just build`: Build the CLI binary
- `just test`: Run tests
- `just check`: Format, lint, and test (runs automatically on commit)

## Code Style

- **Language**: Go
- **Naming**: Kebab-case with numeric prefix for specs and files
- **Documentation**: Prioritize clarity; explain "why" in specs; code shows "how"
- **Types**: Use `any` instead of `interface{}` (Go 1.18+)
- **Formatting**: Code must pass `go fmt` and `go vet`
- **Commits**: Use conventional commits (feat:, fix:, test:, refactor:, etc.)

## CLI Tools (in development)

- `specture setup`: Initialize Specture in a repo
- `specture new`: Create new spec with template
- `specture validate`: Validate spec files
