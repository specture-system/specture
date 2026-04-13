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

The `Draft Release` workflow requires a GitHub App installation token. Configure it by:

1. Creating a GitHub App for release automation.
2. Granting the app `Contents: Read and write` and `Pull requests: Read and write`.
3. Installing the app on this repository.
4. Saving the app's client ID as the repository variable `RELEASE_BOT_CLIENT_ID`.
5. Saving the app's private key as the repository secret `RELEASE_BOT_PRIVATE_KEY`.

The workflow mints a short-lived installation token on each run, so there is no long-lived PAT to rotate.
