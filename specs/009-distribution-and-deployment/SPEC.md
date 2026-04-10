---
number: 9
status: completed
author: Addison Emig
creation_date: 2025-01-22
---

# Distribution and Deployment

Set up automated release infrastructure for the Specture CLI, enabling Linux builds and GitHub releases.

## Goals

- Enable users to install Specture via GitHub releases
- Support Linux `amd64` and `arm64` release builds
- Automate the release process with a single workflow trigger
- Expose build version information through `specture --version`
- Use git tags as the release version source
- Keep `flake.nix` focused on build metadata instead of release version truth

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
  - The output should use the format `v0.3.0 (abc1234)` for tagged releases and `dev (abc1234)` for development builds
- Chosen: Use a short commit hash in version output
  - Seven characters is compact and familiar in CLI output
- Considered: Leaving version information out of the CLI
  - Simpler command surface
  - Gives users no direct way to confirm which build they installed

### Release Trigger

- Chosen: Publish releases from annotated git tag pushes
  - The tag is the source of truth for the release version
  - Tag-driven releases keep the build and published artifact aligned
- Considered: Manual workflow dispatch
  - More flexible for ad hoc publishing
  - Easier to misfire and weaker as a versioning contract

### Release Tooling

- Chosen: GoReleaser
  - Standardizes binaries, archives, checksums, and GitHub Releases
  - Fits the single-trigger release workflow well
- Considered: Custom GitHub Actions
  - More flexible, but requires hand-rolling packaging and publishing steps

### Release Artifacts

- Chosen: Linux binaries, archives, and checksums
  - Covers `amd64` and `arm64`
  - Provides a conventional GitHub Releases download set
- Considered: Binaries only
  - Simpler, but less polished for distribution and verification
- Considered: Signatures and provenance in the initial release
  - Useful later, but adds key-management and workflow scope

### Platform Scope

- Chosen: Linux-only for the initial release
  - Start with `amd64` and `arm64`
  - Keep macOS and Windows as follow-up work
- Considered: Cross-platform releases in the initial release
  - Broader distribution coverage
  - Significantly increases release workflow scope

### Nix Scope

- Chosen: Build metadata only
  - `flake.nix` should remain a build recipe and surface dev/commit identity
  - Release tags should not be baked into the flake
- Considered: Release-aligned packaging
  - Keeps version metadata close to the release
  - Reintroduces version-bump coordination

### Prerelease Tags

- Chosen: Stable tags only for the initial release
  - Keeps the first release pipeline simple
  - Leaves RC and beta flows for later
- Considered: Supporting prerelease tags now
  - Useful for staged rollouts
  - Adds release-channel logic to the initial spec
