This project uses the Specture System. Read `specs/README.md` to learn how the system works, then update {{if .IsClaudeFile}}CLAUDE.md{{else}}AGENTS.md{{end}} with basic information about the system.

Below is a template to include in {{if .IsClaudeFile}}CLAUDE.md{{else}}AGENTS.md{{end}}. Modify as required for this project or for the AI agent workflow you are integrating with.

---

## Specture System

This project uses the Specture System for managing specifications and design documents. When the user asks about planned features, architectural decisions, or implementation details, refer to the `specs/` directory in the repository. Each spec file (specs/NNN-name.md) contains:

- Design rationale and decisions
- Task lists for implementation
- Requirements and acceptance criteria

The `specs/` directory also contains `README.md` with complete guidelines on how the spec system works.

Important guidelines for AI agents and automation:

- Do not edit spec files without explicit permission from a human reviewer. Prompt the user before making changes to any `specs/` file.
- When an agent is used to create or update specs, prefer non-interactive CLI usage patterns described in `specs/README.md`.
  - Use `--title` when supplying a spec body via stdin: `cat spec.md | specture new --title "My Spec"`.
  - Piping a full spec body requires `--title` and implies `--no-editor` (the editor will not be opened).
  - A single-line stdin is treated as the spec title (useful for simple scripting): `echo "Quick Title" | specture new`.
- For repository setup automation, use `specture setup --yes` to skip interactive confirmation. To request AGENTS.md / CLAUDE.md updates even when the files are missing, use `--update-agents` or `--update-claude`.

Be sure to prompt the user for explicit permission before editing the design in any spec file.

When implementing a spec, YOU MUST follow this workflow for each task:

1. Complete a single task from the task list
2. Update the spec file by changing `- [ ]` to `- [x]` for that task
3. Commit both the implementation and spec update together with a conventional commit message (e.g., `feat: implement feature X`)
4. Push the changes

This keeps the spec file as a living document that tracks implementation progress, with each task corresponding to one commit.

---
