# Specs

Specs are design documents that describe planned changes to this project. Each spec file in `specs/` contains the design rationale, decisions, and a task list for implementation. Spec numbers are stored in the YAML frontmatter `number` field. Filenames are slugs (e.g., `status-command.md`).

When writing specs:

- Do not number section headers (`##`/`###`).
- Treat each `###` section in `## Task List` as one pull request.
- Treat each task checkbox as one atomic commit.
- Use inline markdown links with correct relative paths for any cross-spec mentions (for example, `[Status command](status-command.md)`).

For full documentation on the spec system, workflow, and file format, see the [Specture System](https://github.com/specture-system/specture) repository. Run `specture help` for CLI usage.

This project uses {{.ContributionType}}s for collaboration on specs.
