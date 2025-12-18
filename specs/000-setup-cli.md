# Implement Basic CLI

Implement a basic CLI that makes it convenient to use the Specture System in any repoistory.

## Tools

### Setup Project

`specture setup`

Alias: `update`

This makes it easy to add the Specture System to any git repository.

The tool will exit early if it doesn't detect a git repository in the current directory, or if the repository has uncommitted changes. Users will be prompted to verify the generated changes before committing. The tool will _not_ automatically commit.

The tool looks at the git remotes to determine which forge they are using. If there are no remotes, prompt the user to find out which forge.

- If they are using GitLab, the generated files should refer to "merge requests"
- Otherwise, the generated files should refer to "pull requests"

Things that will be generated:

- `specs/` directory for spec files
- `specs/README.md` with the spec guidelines

The tool should automatically detect if the repo has the following files:

- `AGENTS.md`
- `CLAUDE.md`

For each file, the CLI should prompt the user if they want to update that file. If yes, then the CLI should give them a prompt to copy and paste into their agent. The prompt will be something to the effect of "This project uses the Specture System. Read specs/README.md to learn about the system, then update AGENTS.md with basic information for agents. The agents should reference the file when they need more information about the system."

A `--dry-run` flag will allow users to preview all changes without modifying any files or creating commits. This mode will be particularly useful for automated testing within Specture itself, ensuring the CLI behaves correctly across different repository configurations.

The tool will be defensive against accidentally overwriting existing spec files, protecting user-created specifications. However, it will freely replace `specs/README.md` to ensure repositories stay up-to-date as the Specture System evolves.

Users can run the tool in repos that already have the Specture System installed to pull in the latest guidelines and improvements.

### New Spec

`specture new`

alias: `n`

This makes it easy to add new specs.

It should automate all the following:

- Create branch
- Create file based on basic template
- Open file in user's editor

### Validate Spec

`specture validate`

alias: `v`

This makes it easy to validate specs to make sure they follow the Specture System.

It should check the following:

- Valid frontmatter
  - Valid status
- Valid description
- Valid task list

It should be possible to validate one specific spec or all the specs.

## Design Decisions

### Programming Language

- Chosen: go
  - Easy to build standalone binary
  - Good CLI tooling
  - Fast
- Considered: bash
  - Designed for scripting
  - Hard to maintain
  - Hard to implement complex features
  - Slow

## Task List

TBD
