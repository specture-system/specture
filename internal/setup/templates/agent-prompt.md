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

When implementing a spec, follow this workflow for each task:

1. Complete a single task from the task list
2. Commit the implementation with a conventional commit message (e.g., `feat: implement feature X` or `test: add tests for feature Y`)
3. Update the spec file by changing `- [ ]` to `- [x]` for that task
4. Commit the spec update with message `spec: mark task as complete`
5. Push the changes

This keeps the spec file as a living document that tracks implementation progress, with each task corresponding to one commit.

---
