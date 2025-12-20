# AGENTS.md

## Specture System

This project uses the Specture System for managing specifications and design documents. When you ask about planned features, architectural decisions, or implementation details, refer to the specs/ directory in the repository. Each spec file (specs/NNN-name.md) contains:

- Design rationale and decisions
- Task lists for implementation
- Requirements and acceptance criteria

The specs/ directory also contains README.md with complete guidelines on how the spec system works.

**Important**: Before editing the design in any spec file, prompt the user for explicit permission.

When implementing a spec, follow this workflow for each task:

1. Complete a single task from the task list
2. Update the spec file by changing `- [ ]` to `- [x]` for that task
3. Commit both the implementation and spec update together with a conventional commit message (e.g., `feat: implement feature X`)
4. Push the changes

This keeps the spec file as a living document that tracks implementation progress, with each task corresponding to one commit.

### Key Concepts

- **Specs as living documents**: Specs are continually improved during design and implementation, but left static after completion
- **Scope**: Specs cover planned changes—new features, major refactors, redesigns, tooling improvements. Use the issue tracker for bugs
- **Status workflow**: draft → approved → in-progress → completed (or rejected)
- **Precedence**: Higher-numbered specs take precedence when conflicts arise. Once a spec is completed, treat it as historical record; don't retroactively update it (fix only typos/factual errors)
- **Task organization**: Tasks are grouped into logical sections (e.g., Foundation, Core Implementation, Polish and Documentation)
- **File naming**: Numeric prefix with kebab-case (e.g., `000-mvp.md`, `013-refactor-database.md`). Higher numbers have higher precedence

### Directory Structure

- **`specs/`**: Markdown spec files for planned changes (features, refactors, improvements)
- **`specs/README.md`**: Complete spec guidelines, workflow, and best practices
- **`cmd/`**: CLI command definitions and command-specific orchestration
- **`internal/prompt/`**: User interaction utilities (confirmations, prompts, template display)
- **`internal/fs/`**: File system operations
- **`internal/git/`**: Git repository operations
- **`internal/setup/`**: Setup command logic
- **`AGENTS.md`**: This file, for agentic coding tools

## Development Environment

### Setup

Before starting work, ensure pre-commit hooks are installed:

```bash
pip install pre-commit
pre-commit install
```

The pre-commit hook is **required** and automatically runs before every commit. It executes `just check`, which:

- Formats code with `go fmt`
- Runs linting with `go vet`
- Runs the full test suite

**You do not need to manually run these checks.** The pre-commit hook runs them automatically when you attempt to commit. If any checks fail, the commit is blocked and you must fix the issues. If code formatting changes are needed, `go fmt` will modify the files automatically—simply stage the changes and attempt to commit again.

### Build and Test Commands

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

## Code Style and Organization

- **Language**: Go
- **Naming**: Kebab-case with numeric prefix for specs and files
- **Documentation**: Prioritize clarity; explain "why" in specs; code shows "how"
- **Types**: Use `any` instead of `interface{}` (Go 1.18+)
- **Formatting**: Code must pass `go fmt` and `go vet`
- **Commits**: Use conventional commits (feat:, fix:, test:, refactor:, etc.)
- **File organization**: Core functions at top, helper functions at bottom

### Helper Functions and Code Organization

When adding functionality:

1. **Check for existing helpers first**: Before writing new utility functions, search the codebase for similar functionality. Look in `internal/` packages for utilities that might already exist or could be extended.

2. **Place helpers in the correct packages**:
   - `internal/prompt/`: User interaction utilities (confirmations, prompts, template display)
   - `internal/fs/`: File system operations
   - `internal/git/`: Git repository operations
   - `internal/setup/`: Setup command logic
   - `cmd/`: Only CLI command definitions and command-specific orchestration

   **Do NOT** put reusable helper functions in `cmd/` files—they belong in `internal/` packages.

3. **Extract and generalize**: If writing a function in a command file that could be useful elsewhere (e.g., showing a template to a user), move it to the appropriate `internal/` package with unit tests.

4. **Write tests alongside helpers**: All utility functions in `internal/` packages must have corresponding unit tests.

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
