This project uses the Specture System. Read `specs/README.md` to learn how the system works, then update {{if .IsClaudeFile}}CLAUDE.md{{else}}AGENTS.md{{end}} with basic information about the system.

Below is a template to include in {{if .IsClaudeFile}}CLAUDE.md{{else}}AGENTS.md{{end}}. Modify as required for this project or for the AI agent workflow you are integrating with.

---

## Specture System

This project uses the Specture System for managing specifications and design documents. When the user asks about planned features, architectural decisions, or implementation details, refer to the `specs/` directory in the repository. Each spec file (specs/NNN-name.md) contains:

- Design rationale and decisions
- Task lists for implementation
- Requirements and acceptance criteria

The `specs/` directory also contains `README.md` with complete guidelines on how the spec system works.

### Implementation Workflow

**CRITICAL**: When implementing a spec, each task MUST be exactly one commit containing both the implementation AND the spec file update (change `- [ ]` to `- [x]`). Do NOT commit implementation changes without the corresponding spec update in the same commit.

**Important**: Only edit spec files to mark tasks as complete during implementation. Do not retroactively update completed specs or modify design decisions without explicit user permission.

### CLI Usage for AI Agents

**IMPORTANT**: Always use non-interactive flags when running `specture` commands. The default interactive mode will hang waiting for user input and cause your workflow to fail.

Use these flags:
- `specture new --title "Spec Title"` (non-interactive spec creation)
- `specture setup --yes` (non-interactive setup)
- Pipe spec content: `cat body.md | specture new --title "Spec Title"`

Run `specture --help` and `specture <command> --help` to learn about all available flags and options. See `specs/README.md` for complete non-interactive CLI examples.

---
