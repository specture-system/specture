---
number: 9
status: draft
author: Addison Emig
creation_date: 2025-01-22
---

# Distribution and Deployment

Set up automated release infrastructure for the Specture CLI, enabling cross-platform builds and GitHub releases.

## Goals

- Enable users to install Specture via GitHub releases
- Support Linux, macOS, and Windows platforms
- Automate the release process with a single workflow trigger
- Keep flake.nix in sync with releases for Nix users
- Expose build version information through `specture --version`
- Use git tags as the release version source without hardcoding release values in `flake.nix`

## Design Decisions

### Version Source

- Chosen: Inject version metadata at build time
  - Tagged release builds use the git tag as the version string
  - Development builds can fall back to `dev` plus commit information
  - This keeps `flake.nix` focused on building and avoids manual version bumps
- Considered: Hardcoding the version in `flake.nix`
  - Simple for one-off builds
  - Requires manual edits or version-bump commits for every release
  - Creates drift risk between tags and packaged artifacts
- Considered: Deriving the version from remote tags inside Nix
  - Avoids committed version bumps
  - Makes local and offline builds less predictable

### CLI Version Flag

- Chosen: Add a root-level `--version` flag
  - The flag should print the build version and exit without running other commands
  - The output may include commit information for development builds
- Considered: Leaving version information out of the CLI
  - Simpler command surface
  - Gives users no direct way to confirm which build they installed
