---
status: approved
author: Addison Emig
creation_date: 2026-01-06
approved_by: Bennett Moore
approval_date: 2026-01-11
---

# Non-interactive Input

The current specture CLI requires interactive input to create new specs. This makes it difficult to use from automated workflows, for example, AI agents.

We should improve all the commands implemented in [spec #000](/specs/000-basic-cli.md) with optional flags that enable full configuration without requiring any interactive input.

## Design Decisions

### Explicit flags instead of generic `--yes`

- Chosen: Explicit flags for each prompt (`--title`, `--update-agents`, etc.)
  - Gives precise control over each interactive input
  - Safer for automation—no accidental "yes to everything"
  - Self-documenting command lines
- Considered: Generic `--yes` / `-y` flag
  - Common convention (apt, npm, pacman)
  - Less control—blindly accepts all prompts
  - Could lead to unintended side effects

### Confirmation behavior differs by command

- `specture setup`: Requires explicit `--yes` / `-y` to skip confirmation
  - Makes significant changes (creates directories, files)
  - Follows glab convention for destructive operations
  - Safer for users who accidentally run command
- `specture new`: Skips confirmation when `--title` is provided
  - Providing `--title` signals clear intent to proceed
  - Lighter-weight operation (single file, single branch)
  - Reduces flag count for automation

### Update flags override file detection

- Chosen: `--update-agents` and `--update-claude` trigger update logic even if files don't exist
  - User might be setting up a brand new repo without these files yet
  - Explicit flag signals clear intent
  - Allows automation to request update prompt for files that will be created
- Considered: Only show update prompt if file already exists
  - Current behavior in interactive mode
  - Too restrictive for new repo setup workflows

### Spec content from stdin

- Chosen: Pipe content replaces entire spec body (frontmatter auto-generated)
  - AI agents can generate full spec content programmatically
  - Frontmatter (status, author, date) still managed by CLI
  - Works naturally with shell pipes: `cat content.md | specture new --title "My Spec"`
  - Automatically implies `--no-editor` (content already provided)
- Considered: Pipe content for description section only
  - Too limiting—agents want to write full specs
  - Would require parsing/merging content

## Task List

### Setup Command (`specture setup`)

- [x] Write tests for `--yes` / `-y` flag behavior
- [x] Add `--yes` / `-y` flag to skip confirmation prompt
- [x] Write tests for `--update-agents` flag behavior
- [x] Add `--update-agents` flag to show update prompt (even if file doesn't exist)
- [x] Write tests for `--no-update-agents` flag behavior
- [x] Add `--no-update-agents` flag to skip AGENTS.md update
- [x] Write tests for `--update-claude` flag behavior
- [x] Add `--update-claude` flag to show update prompt (even if file doesn't exist)
- [x] Write tests for `--no-update-claude` flag behavior
- [x] Add `--no-update-claude` flag to skip CLAUDE.md update

### New Command (`specture new`)

- [x] Write tests for `--title` / `-t` flag behavior
- [x] Add `--title` / `-t` flag for spec title
- [x] Write tests for `--no-editor` flag behavior
- [x] Add `--no-editor` flag to skip opening editor
- [x] Write tests for stdin content piping (including auto `--no-editor`)
- [x] Write tests for early exit when stdin is piped but `--title` not provided
- [x] Detect and read spec content from stdin pipe
- [x] Exit early with error when stdin is piped but `--title` not provided
- [x] Write tests for skipping confirmation when `--title` is provided
- [x] Skip confirmation prompt when `--title` is provided

### Documentation

Well written command help text is critical for both humans and AI agents to understand how to use the specture CLI.  
Describing the various nuances is essential for the best CLI user experience.

- [x] Update `specture setup` help text with new flags and behaviors
  - [x] Document that `--yes` is required to skip confirmation prompt
  - [x] Document that `--update-agents` / `--update-claude` trigger update prompt even when files aren't detected
- [x] Update `specture new` help text with new flags and behaviors
  - [x] Document that `--title` is required when piping content to stdin
  - [x] Document that `--no-editor` is automatically implied when content is piped to stdin
  - [x] Document that confirmation is skipped when `--title` is provided
- [ ] Add non-interactive usage examples to CLI documentation
- [ ] Update `internal/templates/files/agent-prompt.md`
