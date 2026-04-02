# Specs

Specs are design documents that describe planned changes to this project. Each spec file in `specs/` contains the design rationale and decisions. Track implementation progress in a sibling `PROGRESS.md` file next to the spec when needed, and keep it ephemeral and untracked. Spec numbers are stored in the YAML frontmatter `number` field. Filenames are slugs (e.g., `status-command.md`).

When writing specs:

- Do not number section headers (`##`/`###`).
- Use inline markdown links with correct relative paths for any cross-spec mentions (for example, `[Status command](status-command.md)`).
- Keep implementation progress in sibling `PROGRESS.md` files, not in the spec itself.

For full documentation on the spec system, workflow, and file format, see the [Specture System](https://github.com/specture-system/specture) repository. Run `specture help` for CLI usage.

This project uses pull requests for collaboration on specs.
