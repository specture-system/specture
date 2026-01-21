This project uses the Specture System. Read `specs/README.md` to learn how the system works, then update {{if .IsClaudeFile}}CLAUDE.md{{else}}AGENTS.md{{end}} with basic information about the system.

Below is a template to include in {{if .IsClaudeFile}}CLAUDE.md{{else}}AGENTS.md{{end}}. Modify as required for this project or for the AI agent workflow you are integrating with.

---

## Specture System

This project uses the Specture System for managing specifications and design documents. When the user asks about planned features, architectural decisions, or implementation details, refer to the `specs/` directory in the repository. Each spec file (specs/NNN-name.md) contains:

- Design rationale and decisions
- Task lists for implementation
- Requirements and acceptance criteria

The `specs/` directory also contains `README.md` with complete guidelines on how the spec system works.

**Important**: Do not edit spec files without explicit user permission.

**CRITICAL**: When implementing a spec, each task MUST be exactly one commit containing both the implementation AND the spec file update (change `- [ ]` to `- [x]`). Do NOT commit implementation changes without the corresponding spec update in the same commit.

For non-interactive CLI usage (`specture new`, `specture setup`), see `specs/README.md`.

---
