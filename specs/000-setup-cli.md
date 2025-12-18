# Implement Setup CLI

## Description

Implement a basic CLI tool to make it convenient to add the Specture System to any git repository. The CLI will be written in Go and prompt the user for configuration options like git forge.

This tool will streamline the onboarding process for teams wanting to adopt the Specture System, automating the creation of necessary directories, documentation, and configuration files. The CLI will set up:

- `specs/` directory for spec files
- `specs/README.md` with the spec guidelines
- `AGENTS.md` file (adding a Specture System section without overwriting existing content)

The CLI will exit early if it doesn't detect a git repository in the current directory, or if the repository has uncommitted changes. Users will be able to disable individual changes they don't want, and will always be prompted to verify the proposed changes before committing. This interactive approach ensures users maintain full control over what gets added to their repository while reducing manual setup effort.

The CLI will be defensive against accidentally overwriting existing spec files, protecting user-created specifications. However, it will freely replace `specs/README.md` to ensure repositories stay up-to-date as the Specture System evolves. The CLI will support updating repositories that already have the Specture System installed, making it easy to pull in the latest guidelines and improvements.

## Design Decisions

TBD

## Task List

TBD
