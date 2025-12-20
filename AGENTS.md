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

**Always use `just` recipes for development tasks.** Do NOT run `go` commands directly.

Use `just` to run development tasks. See `justfile` for available recipes. Common commands:

- `just build`: Build the CLI binary
- `just test`: Run tests
- `just check`: Format, lint, and test (runs automatically on commit)
- `just fmt`: Format code
- `just lint`: Run linters
- `just install`: Install the CLI locally
- `just clean`: Clean build artifacts

**Important**: The project requires `CGO_ENABLED=0` for builds, which is configured in the `justfile`. Running `go` commands directly without this flag will fail in the Nix environment.

## Code Style

- **Language**: Go
- **Naming**: Kebab-case with numeric prefix for specs and files
- **Documentation**: Prioritize clarity; explain "why" in specs; code shows "how"
- **Types**: Use `any` instead of `interface{}` (Go 1.18+)
- **Formatting**: Code must pass `go fmt` and `go vet`
- **Commits**: Use conventional commits (feat:, fix:, test:, refactor:, etc.)
- **File organization**: Core functions at top, helper functions at bottom

## CLI Tools (in development)

- `specture setup`: Initialize Specture in a repo
- `specture new`: Create new spec with template
- `specture validate`: Validate spec files

## Spec Editing Safety

Spec files under `specs/` are long-term design documents. Follow this strict workflow:

1. **Ask for confirmation before editing any spec file**. State the exact file path and change summary. Example:
   ```
   I plan to update `specs/001-new-data-format` to change the 'Blank entries representation' section. Reply 'yes' to apply.
   ```

2. **After receiving approval**, stage files and present a one-line commit message for explicit confirmation before committing.

3. **Do not commit spec changes without explicit user approval**. If uncertain whether something is a transient implementation choice or long-term spec decision, ask the user.

This ensures specs remain authoritative design documents and changes are intentional.

## GitHub Workflow

**Creating issues and PRs:**
```bash
# View issue details
gh issue view <number>

# Create PR with conventional commit format (required)
gh pr create --title "feat: add new feature" --body "Description of changes"
```

**PR Title Format**: PRs are squashed on merge to main, so PR titles become commit messages. Use conventional commit format:
- `feat:` for features
- `fix:` for bug fixes
- `refactor:` for refactoring
- `docs:` for documentation
- `test:` for test changes

## Nix Development

Specture uses Nix flakes for reproducible builds.

**Updating vendorHash when go.mod changes:**

1. Set `vendorHash = "";` in `flake.nix` (empty string)
2. Run `nix build 2>&1 | grep -E "(specified|got):"`
3. Nix will show the correct hash:
   ```
   specified: sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
   got: sha256-ZknVM8bMM0kLIbuV4Bv4XsbgtyhlKyP7p2AVOE1k0GA=
   ```
4. Copy the `got:` hash and update `vendorHash` in `flake.nix`
5. Run `nix build` again to verify it works
6. **Do not run `go mod vendor`** — the vendor directory should remain empty
