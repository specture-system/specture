---
status: in-progress
author: Addison Emig
creation_date: 2025-12-18
approved_by: Addison Emig
approval_date: 2025-12-19
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

### Markdown and YAML Parsing

- Chosen: goldmark with frontmatter extension
  - **github.com/yuin/goldmark** - CommonMark compliant markdown parser
  - **github.com/abhinav/goldmark-frontmatter** - Frontmatter extension for goldmark
  - Unified approach for both markdown structure and YAML frontmatter parsing
  - Proper AST for robust validation (headings, task lists, sections)
  - Future-proof for more complex markdown parsing requirements
  - Industry standard for Go markdown processing
  - Extensible for additional validation rules
- Considered: gopkg.in/yaml.v3 + simple string/regex parsing
  - Would work for current basic validation needs
  - Separate approaches for frontmatter vs markdown
  - Manual parsing less robust and harder to extend
  - Would require migration to proper parser later
- Considered: Simple string/regex parsing only
  - No dependencies
  - Too fragile for reliable validation
  - Difficult to maintain and extend

### Build and Release Tooling

- Chosen: GoReleaser
  - Industry standard for Go CLI releases
  - Automatic GitHub release generation with one command
  - Cross-platform builds (Linux, macOS, Windows) configured in one file
  - Automatic changelog generation
  - Checksums and archive creation
  - Simple GitHub Actions integration
  - Consolidates build/release logic that would otherwise require multiple scripts
- Considered: Manual builds with Makefile + gh CLI
  - Simpler initial setup
  - Requires manual build scripts for each platform
  - Manual release creation and binary uploads
  - Gets tedious with frequent releases
  - More error-prone

### Development Task Runner

- Chosen: just (justfile)
  - Clean, modern syntax (no Make's quirks)
  - Designed specifically for running commands (not dependency management)
  - Better error messages than Make
  - Cross-platform support
  - Easy to read and maintain
- Considered: Makefile
  - More ubiquitous (installed everywhere)
  - Complex syntax with tabs/spaces issues
  - Designed for build dependencies, not task running
  - Less intuitive for simple command aliases

## Task List

### Project Setup

- [x] Initialize Go module and project structure
- [x] Set up Cobra CLI framework
- [x] Set up testing infrastructure (framework, helpers, test fixtures)
- [x] Configure justfile for local development commands

### Core Infrastructure

- [x] Write unit tests for git repository detection
- [x] Implement git repository detection (using os/exec)
- [x] Write unit tests for uncommitted changes check
- [x] Implement uncommitted changes check (using os/exec)
- [x] Write unit tests for forge identification
- [x] Implement git remote detection and forge identification (GitLab vs others)
- [x] Write unit tests for terminology detection
- [x] Create utility for terminology detection ("merge request" vs "pull request")
- [x] Write unit tests for file system utilities
- [x] Create file system utilities (safe read/write, directory creation)
- [x] Write unit tests for prompt system (with mocked input)
- [x] Implement user prompt/confirmation system
- [x] Write unit tests for template utilities
- [x] Create text/template-based markdown file generation utilities

### Setup Command (`specture setup`)

- [x] Implement basic command structure and aliases (`setup`, `update`)
- [x] Write integration tests for setup command preconditions
- [x] Add git repository validation (exit if not a git repo)
- [x] Add uncommitted changes check (exit if dirty working tree)
- [x] Implement forge detection logic
- [x] Write tests for dry-run mode (no file modifications)
- [x] Add `--dry-run` flag support
- [x] Write integration tests for setup command file generation
- [x] Create `specs/` directory generation
- [x] Create `specs/README.md` template with forge-appropriate terminology
- [x] Implement `specs/README.md` generation/update logic
- [x] Write tests for AGENTS.md/CLAUDE.md detection
- [x] Implement `AGENTS.md` detection and update prompt
- [x] Implement `CLAUDE.md` detection and update prompt
- [x] Write tests for overwrite protection
- [x] Add protection against overwriting existing spec files
- [x] Write integration tests for complete setup workflow
- [x] Implement user confirmation flow before making changes
- [x] Add comprehensive error handling and user-friendly messages

### New Spec Command (`specture new`)

- [ ] Implement basic command structure and alias (`new`, `n`)
- [ ] Write tests for spec template generation
- [ ] Create spec file template with YAML frontmatter (using text/template)
- [ ] Write tests for spec numbering logic
- [ ] Implement automatic spec numbering (find next available number)
- [ ] Write tests for branch creation (with test git repos)
- [ ] Implement branch creation with appropriate naming (using git CLI via os/exec)
- [ ] Add user prompt for spec title/description
- [ ] Write integration tests for complete new spec workflow
- [ ] Implement file creation from template
- [ ] Implement editor detection and opening (respect $EDITOR)
- [ ] Write tests for error handling scenarios
- [ ] Add error handling for edge cases (no git, existing file, etc.)

### Validate Command (`specture validate`)

- [ ] Implement basic command structure and alias (`validate`, `v`)
- [ ] Write tests for spec parsing (valid and invalid specs)
- [ ] Implement goldmark-based spec parser with frontmatter extension
- [ ] Write tests for frontmatter validation
- [ ] Add frontmatter validation (required fields present)
- [ ] Write tests for status validation
- [ ] Add status field validation (draft/approved/in-progress/completed/rejected)
- [ ] Write tests for description validation
- [ ] Implement description section validation (using goldmark AST)
- [ ] Write tests for task list validation
- [ ] Implement task list detection and validation (using goldmark AST)
- [ ] Write integration tests for both validation modes
- [ ] Add single-spec validation mode (by file path or number)
- [ ] Add all-specs validation mode
- [ ] Write tests for error messages and summary output
- [ ] Implement clear, actionable error messages for validation failures
- [ ] Add summary output (X of Y specs valid)

### Documentation

- [ ] Create CLI usage documentation
- [ ] Add command-line help text for all commands

### Distribution & Deployment

- [ ] Configure GoReleaser to inject version via ldflags at build time
- [ ] Implement version display in main.go (read injected build flag, with fallback for local builds)
- [ ] Write tests for version display
- [ ] Add `--version` flag to root command
- [ ] Configure GoReleaser for multi-platform builds (Linux, macOS, Windows)
- [ ] Create manually-triggered GitHub Actions workflow for releases
- [ ] Workflow accepts version input parameter
- [ ] Workflow updates flake.nix with version
- [ ] Workflow commits changes with conventional commit message
- [ ] Workflow creates annotated git tag matching version
- [ ] Workflow runs GoReleaser for cross-platform builds
- [ ] Workflow creates GitHub release with artifacts
- [ ] Create installation instructions
- [ ] Document how to trigger release workflow (GitHub UI or CLI)

### Cross-Platform Testing

- [ ] Test on different repository configurations (GitLab, GitHub, no remote)
- [ ] Test on different platforms (Linux, macOS, Windows if available)
