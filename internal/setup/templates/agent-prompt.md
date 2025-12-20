This project uses the Specture System. Read specs/README.md to learn about how the system works, then update {{if .IsClaudeFile}}CLAUDE.md{{else}}AGENTS.md{{end}} with basic information about the system.

Below is a template to include in {{if .IsClaudeFile}}CLAUDE.md{{else}}AGENTS.md{{end}}. Modify as required for this project.

---

## Specture System

This project uses the Specture System for managing specifications and design documents. When the user asks about planned features, architectural decisions, or implementation details, refer to the specs/ directory in the repository. Each spec file (specs/NNN-name.md) contains:

- Design rationale and decisions
- Task lists for implementation
- Requirements and acceptance criteria

The specs/ directory also contains README.md with complete guidelines on how the spec system works.

Be sure to prompt the user for explicit permission before editing the design in any spec file.

When implementing a spec, check off each item in the task list as you go.

---
