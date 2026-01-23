# 🏗️ Specture

> Spec-driven software architecture system

Specture is a spec-driven software architecture and project management system.

### The Concept

- **Specs for Planned Changes**: New features, major refactors, redesigns, and tooling improvements are added as markdown files in `specs/`, with discussion happening in the pull request that adds the spec
- **Issues for Bugs Only**: The issue tracker is for bugs - cases where the software doesn't match what's described in the specs
- **AI-Friendly Workflow**: Designed to work seamlessly with AI agents that help build and maintain your codebase
- **Small Team Focus**: Built for teams where lightweight, document-driven planning makes sense

See [specs/README.md](/specs/README.md) for a full description of the Specture System.

### Status

This project is in its early stages. Documentation and tooling are a work-in-progress.

## Pre-commit Hooks

To enable Specture's pre-commit hooks in your project, add this to your `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/specture-system/specture
    rev: v0.0.1 # Use the latest release tag
    hooks:
      - id: validate-specs
```

This will run `specture validate` to ensure specs conform to the Specture format.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for the full text.

All contributions must be submitted with the understanding that they will be released under the MIT License.

## Contributing

Issues are for bugs only. For features, refactors, or other changes, submit a pull request adding a spec file to `specs/`.
