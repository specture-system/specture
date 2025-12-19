---
status: draft
author: Addison Emig
creation_date: 2025-12-18
---

# Implement Basic CLI

Implement a basic CLI that makes it convenient to use the Specture System in any repository.

## Tools

### Setup Project

`specture setup`

Alias: `update`

This makes it easy to add the Specture System to any git repository.

The tool will exit early if it doesn't detect a git repository in the current directory, or if the repository has uncommitted changes. Users will be prompted to verify the generated changes before committing. The tool will _not_ automatically commit.

The tool looks at the git remotes to determine which forge they are using. If there are no remotes, prompt the user to find out which forge.

- If they are using GitLab, the generated files should refer to "merge requests"
- Otherwise, the generated files should refer to "pull requests"

Things that will be generated:

- `specs/` directory for spec files
- `specs/README.md` with the spec guidelines

The tool should automatically detect if the repo has the following files:

- `AGENTS.md`
- `CLAUDE.md`

For each file, the CLI should prompt the user if they want to update that file. If yes, then the CLI should give them a prompt to copy and paste into their agent. The prompt will be something to the effect of "This project uses the Specture System. Read specs/README.md to learn about the system, then update AGENTS.md with basic information for agents. The agents should reference the file when they need more information about the system."

A `--dry-run` flag will allow users to preview all changes without modifying any files or creating commits. This mode will be particularly useful for automated testing within Specture itself, ensuring the CLI behaves correctly across different repository configurations.

The tool will be defensive against accidentally overwriting existing spec files, protecting user-created specifications. However, it will freely replace `specs/README.md` to ensure repositories stay up-to-date as the Specture System evolves.

Users can run the tool in repos that already have the Specture System installed to pull in the latest guidelines and improvements.

### New Spec

`specture new`

alias: `n`

This makes it easy to add new specs.

It should automate all the following:

- Create branch
- Create file based on basic template
- Open file in user's editor

### Validate Spec

`specture validate`

alias: `v`

This makes it easy to validate specs to make sure they follow the Specture System.

It should check the following:

- Valid frontmatter
  - Valid status
- Valid description
- Valid task list

It should be possible to validate one specific spec or all the specs.

## Design Decisions

### Programming Language

- Chosen: go
  - Easy to build standalone binary
  - Good CLI tooling
  - Fast
- Considered: bash
  - Designed for scripting
  - Hard to maintain
  - Hard to implement complex features
  - Slow

### CLI Framework

- Chosen: Cobra (github.com/spf13/cobra)
  - Industry standard (used by kubectl, Hugo, GitHub CLI, Docker CLI)
  - First-class support for subcommands and aliases
  - Excellent auto-generated help text
  - Team familiarity from other projects
  - Scales well as the CLI grows
- Considered: urfave/cli
  - Lighter weight than Cobra
  - Simpler API
  - Less feature-rich for complex CLIs
- Considered: Standard library `flag`
  - Zero dependencies
  - No built-in subcommand or alias support
  - Would require significant manual implementation

### Template Engine

- Chosen: `text/template` (standard library)
  - Zero dependencies (standard library)
  - Handles conditionals well (needed for forge-specific terminology)
  - Can embed templates in binary using `go:embed`
  - Standard for Go developers
  - Perfect for markdown generation
- Considered: Simple string replacement (`fmt.Sprintf`, `strings.Replace`)
  - Very simple for basic cases
  - Poor support for conditionals and logic
  - Hard to maintain for complex templates
- Considered: Third-party libraries (pongo2, raymond)
  - Additional dependencies
  - Overkill for our use case

### Git Interaction

- Chosen: Shell out to git CLI using `os/exec`
  - Git is already a requirement (tool exits if not a git repo)
  - Simple implementation for basic operations
  - Respects user's git configuration and hooks
  - Full feature parity with git CLI
  - Only need 4 simple operations: repo check, status check, remote detection, branch creation
  - Easier to debug and test
- Considered: go-git library
  - Pure Go, no git dependency
  - Large dependency (~3MB added to binary)
  - Doesn't respect user's git config/hooks
  - More complex than needed for simple operations
  - Overkill for our use case

## Task List

### Project Setup

- [ ] Initialize Go module and project structure
- [ ] Set up Cobra CLI framework
- [ ] Configure build system and Makefile
- [ ] Set up basic testing infrastructure

### Core Infrastructure

- [ ] Implement git repository detection
- [ ] Implement uncommitted changes check
- [ ] Implement git remote detection and forge identification (GitLab vs others)
- [ ] Create utility for terminology detection ("merge request" vs "pull request")
- [ ] Create file system utilities (safe read/write, directory creation)
- [ ] Implement user prompt/confirmation system
- [ ] Create template engine for generating markdown files

### Setup Command (`specture setup`)

- [ ] Implement basic command structure and aliases (`setup`, `update`)
- [ ] Add git repository validation (exit if not a git repo)
- [ ] Add uncommitted changes check (exit if dirty working tree)
- [ ] Implement forge detection logic
- [ ] Add `--dry-run` flag support
- [ ] Create `specs/` directory generation
- [ ] Create `specs/README.md` template with forge-appropriate terminology
- [ ] Implement `specs/README.md` generation/update logic
- [ ] Implement `AGENTS.md` detection and update prompt
- [ ] Implement `CLAUDE.md` detection and update prompt
- [ ] Add protection against overwriting existing spec files
- [ ] Implement user confirmation flow before making changes
- [ ] Add comprehensive error handling and user-friendly messages

### New Spec Command (`specture new`)

- [ ] Implement basic command structure and alias (`new`, `n`)
- [ ] Create spec file template with YAML frontmatter
- [ ] Implement automatic spec numbering (find next available number)
- [ ] Implement branch creation with appropriate naming
- [ ] Add user prompt for spec title/description
- [ ] Implement file creation from template
- [ ] Implement editor detection and opening (respect $EDITOR)
- [ ] Add error handling for edge cases (no git, existing file, etc.)

### Validate Command (`specture validate`)

- [ ] Implement basic command structure and alias (`validate`, `v`)
- [ ] Implement YAML frontmatter parser
- [ ] Add frontmatter validation (required fields present)
- [ ] Add status field validation (draft/approved/in-progress/completed/rejected)
- [ ] Implement description section validation
- [ ] Implement task list detection and validation
- [ ] Add single-spec validation mode (by file path or number)
- [ ] Add all-specs validation mode
- [ ] Implement clear, actionable error messages for validation failures
- [ ] Add summary output (X of Y specs valid)

### Testing & Documentation

- [ ] Write unit tests for core utilities
- [ ] Write integration tests for `setup` command
- [ ] Write integration tests for `new` command
- [ ] Write integration tests for `validate` command
- [ ] Test dry-run mode thoroughly
- [ ] Create CLI usage documentation
- [ ] Add command-line help text for all commands
- [ ] Test on different repository configurations (GitLab, GitHub, no remote)

### Distribution & Deployment

- [ ] Configure build for multiple platforms (Linux, macOS, Windows)
- [ ] Create installation instructions
- [ ] Set up CI/CD for building releases
- [ ] Create release process documentation
