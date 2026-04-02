---
name: github-workflow
description: Use when working with GitHub issues or pull requests for the Specture repository, including viewing issues, preparing PRs, and choosing conventional-commit titles.
---

# GitHub Workflow

This skill covers the repository's GitHub conventions for issues and pull requests.

Use it when:

- inspecting an issue before implementing work
- preparing a pull request
- choosing a PR title
- checking whether a change should be an issue or a spec-backed PR

## Core Rules

- Issues are for bugs only.
- Features, refactors, and other planned changes should be proposed through specs and pull requests instead of issues.
- Pull request titles must use conventional commit format because PRs are squashed on merge and the title becomes the commit message.
- Use `feat:` for spec-backed implementation work unless the PR is strictly a bug fix, docs change, refactor, or test-only change.

## Commands

View issue details:

```bash
gh issue view <number>
```

Create a pull request:

```bash
gh pr create --title "feat: add new feature" --body "Description of changes"
```

## PR Title Format

Use one of these prefixes:

- `feat:` for features
- `fix:` for bug fixes
- `refactor:` for refactoring
- `docs:` for documentation
- `test:` for test changes

Choose the narrowest accurate prefix. If the work is implementation plus spec-checkbox updates, title the PR by the implementation change, not by the bookkeeping.

## Before Opening a PR

- confirm the branch contains the intended code changes
- run the relevant repo checks
- make sure any required spec updates are included in the same branch
- ensure the PR title is ready to become the squashed commit message
