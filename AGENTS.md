# AGENTS.md

## Guidelines

### Always Do (never ask)

- Use `just` recipes for development tasks instead of running `go` commands directly.
- Prioritize clarity in documentation: explain why in specs, and let code show how.
- Keep core functions near the top of files and helper functions near the bottom.
- Keep reusable logic in the appropriate `internal/` package and keep `cmd/` focused on CLI orchestration.
- Check for existing helpers before adding new ones, and add tests alongside new `internal/` helpers.
- Test public `internal/` helpers, cover edge cases, and prefer table-driven tests when they fit.
- After changing files under `internal/templates/files/`, run `just run update --yes` so the checked-in repo files stay in sync with the templates.
- Use conventional commits for commit messages and PR titles. For PR titles, choose the prefix that matches the primary change; spec bookkeeping should not determine the prefix.

### Ask First (wait for approval)

- Deleting specs, commands, packages, or other files that are not clearly obsolete for the task at hand.
- Changing repo-wide workflow conventions, generated outputs, or documented developer process.

### Never Do (hard stop)

- Running raw `go` commands for build, test, or generation when a `just` recipe should be used.
- Putting reusable helpers in `cmd/` files instead of moving them into `internal/`.

## Specture System

This project uses the [Specture System](https://github.com/specture-system/specture) for managing specs. See the `.agents/skills/specture/` skill for the full workflow, or run `specture help` for CLI usage.
