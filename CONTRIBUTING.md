# Contributing to Specture

## Setup

### With Nix

If you have Nix installed, run:

```bash
nix shell .#default
```

This drops you into a shell with all dependencies pre-installed, including Go, Git, just, and pre-commit.

### Without Nix

Install dependencies and set up pre-commit hooks:

```bash
pip install pre-commit
pre-commit install
```

This ensures code passes all checks (formatting, linting, tests) before committing.

## Development

Use `just` to run development tasks. Run `just --list` to see available recipes.
