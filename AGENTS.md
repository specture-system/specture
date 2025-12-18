# AGENTS.md

## Overview
Specture is a spec-driven software architecture system designed for small teams working with AI agents. It provides a lightweight, document-driven approach to project planning where specs are markdown files in the `specs/` directory.

## Build/Test Commands
No build or test commands yet—project is in early stages. Implementation will be in Go.

## Architecture
- **`specs/`**: Markdown spec files for planned changes (features, refactors, improvements)
- **`specs/README.md`**: Spec guidelines and workflow (do not manually edit; managed by CLI)
- **`AGENTS.md`**: This file, for agentic coding tools

## Key Concepts
- **Specs for Planned Work**: New features, refactors, and improvements are specs in `specs/` (not issues)
- **Issues for Bugs Only**: Bug tracking via issues
- **Spec Structure**: YAML frontmatter (status, author, dates) + Markdown with description and task list
- **Precedence**: Higher spec numbers override lower numbers if they conflict

## Code Style (When Implementation Begins)
- **Language**: Go
- **File Naming**: Kebab-case with numeric prefix: `000-feature.md`
- **Spec Statuses**: `draft`, `approved`, `in-progress`, `completed`, `rejected`
- **Documentation**: Prioritize clarity over cleverness; explain "why" in specs; code shows "how"

## CLI Tools (To Be Implemented)
- `specture setup`: Initialize Specture in a repo
- `specture new`: Create new spec with template
- `specture validate`: Validate spec files

## Important Notes
- The system is inspired by PEP (Python) and BIP (Bitcoin) enhancement proposal systems
- Small team focus with AI agent integration
- See `specs/README.md` for detailed spec guidelines
