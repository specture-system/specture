# AGENTS.md

## Specture System

This project uses the [Specture System](https://github.com/specture-system/specture) for managing specs. See the `.agents/skills/specture/` skill for the full workflow, or run `specture help` for CLI usage.

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
- `just run-dev <args>`: Run the CLI with arguments (e.g., `just run-dev new --help`)
- `just test`: Run tests
- `just check`: Format, lint, and test (runs automatically on commit)
- `just format`: Format code
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

## Testing

- **Test coverage**: Write tests for all public functions in `internal/` packages. Test files live next to implementation files with `_test.go` suffix.
- **Table-driven tests**: Use table-driven tests for testing multiple input/output scenarios:
  ```go
  tests := []struct {
      name     string
      input    string
      expected string
  }{
      {"case 1", "input1", "expected1"},
      {"case 2", "input2", "expected2"},
  }
  for _, tt := range tests {
      t.Run(tt.name, func(t *testing.T) {
          // test logic
      })
  }
  ```
- **Edge cases**: Test edge cases (empty input, missing files, errors, etc.) in separate test runs, not just happy paths.
- **Helper functions**: Create test helper functions (e.g., `InitGitRepo`) in `internal/testhelpers/` to reduce duplication across tests.
- **Mocking**: For file and git operations, use temporary directories (`t.TempDir()`) in tests rather than mocking.
- **Test naming**: Test function names follow pattern `Test<Function><Scenario>` (e.g., `TestCleanup`, `TestCreateBranch`). Use `t.Run()` subtests with descriptive names for different cases.

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
