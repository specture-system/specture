# Contributing to Specture

## Setup

### With Nix

If you have Nix installed, run:

```bash
nix shell .#default
```

This builds the specture binary and drops you into a shell with all dependencies pre-installed, including Go, Git, just, and pre-commit. After making changes, rebuild the CLI with `just build` to test them, then run `./specture` to test your changes.

(Note: `nix shell .` alone provides only the dev tools without building the binary.)

### Without Nix

Install dependencies and set up pre-commit hooks:

```bash
pip install pre-commit
pre-commit install
```

This ensures code passes all checks (formatting, linting, tests) before committing.

## Development

Use `just` to run development tasks. Run `just --list` to see available recipes.

## Releasing

The current release version lives in `VERSION` at the repo root.

To cut a new release:

1. Run the `Draft Release` workflow and choose a major, minor, or patch bump.
2. Review and merge the generated pull request.
3. The `Release` workflow will tag the merged commit and publish the GitHub release.

The `Draft Release` workflow requires a `RELEASE_BOT_TOKEN` secret with permission to create pull requests.

To create that token:

1. Create or choose a GitHub account that will own the bot token.
2. Create a fine-grained personal access token for this repository.
3. Grant it `Contents: Read and write` and `Pull requests: Read and write`.
4. Add the token to the repository secrets as `RELEASE_BOT_TOKEN`.

If your organization prefers GitHub Apps, you can use an app installation token with the same repository permissions instead of a PAT.
