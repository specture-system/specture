# 🏗️ Specture

> Spec-driven software architecture system

## 🚧 Coming Soon

Specture is a spec-driven software architecture and project management system.

### The Concept

- **Specs for Planned Changes**: New features, major refactors, redesigns, and tooling improvements are added as markdown files in `specs/`, with discussion happening in the pull request that adds the spec
- **Issues for Bugs Only**: The issue tracker is for bugs - cases where the software doesn't match what's described in the specs
- **AI-Friendly Workflow**: Designed to work seamlessly with AI agents that help build and maintain your codebase
- **Small Team Focus**: Built for teams where lightweight, document-driven planning makes sense

### Status

This project is in early stages. Documentation and tooling coming soon.

---

**Note**: Issues are for bugs only. For features, refactors, or other changes, submit a pull request adding a spec file to `specs/`.

## Examples: Non-interactive CLI Usage

Add non-interactive examples here to make automation easier.

- Create a new spec by providing the title via flag (skips title prompt and confirmation):

  ```bash
  specture new --title "Add search indexing"
  ```

- Create a new spec from piped content (provide the title flag when piping the full body):

  ```bash
  cat spec-body.md | specture new --title "Automated Spec from AI"
  ```

  Notes:
  - When piping a full spec body, `--title` is required.
  - Piping a body implies `--no-editor` (the editor will not be opened).

- Provide a title via a single-line stdin (useful for simple scripting):

  ```bash
  echo "Quick Title" | specture new
  ```

  The single-line stdin will be treated as the title.

- Setup a repository non-interactively (skip confirmation):

  ```bash
  specture setup --yes
  ```

- Request AGENTS.md update even if the file doesn't exist:

  ```bash
  specture setup --update-agents --yes
  ```

- Preview changes without modifying files:

  ```bash
  specture new --title "Preview" --dry-run
  specture setup --dry-run
  ```

For more detailed documentation and examples, see the `specs/` directory and the CLI help for each command (`specture new --help`, `specture setup --help`).
