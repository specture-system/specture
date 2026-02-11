---
number: 7
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

## Task List

### Version Display

- [ ] Configure GoReleaser to inject version via ldflags at build time
- [ ] Implement version display in main.go (read injected build flag, with fallback for local builds)
- [ ] Write tests for version display
- [ ] Add `--version` flag to root command

### GoReleaser Configuration

- [ ] Configure GoReleaser for multi-platform builds (Linux, macOS, Windows)
- [ ] Configure automatic changelog generation
- [ ] Configure checksums and archive creation

### GitHub Actions Release Workflow

- [ ] Create manually-triggered GitHub Actions workflow for releases
- [ ] Workflow accepts version input parameter
- [ ] Workflow updates flake.nix with version
- [ ] Workflow updates vendorHash in flake.nix
- [ ] Workflow commits changes with conventional commit message
- [ ] Workflow creates annotated git tag matching version
- [ ] Workflow runs GoReleaser for cross-platform builds
- [ ] Workflow creates GitHub release with artifacts

### Documentation

- [ ] Create installation instructions
- [ ] Document how to trigger release workflow (GitHub UI or CLI)
