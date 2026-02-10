---
status: approved
author: Addison Emig
creation_date: 2026-02-05
approved_by: Addison Emig
approval_date: 2026-02-10
---

# Workflow Assistance

The Specture System relies on agents reading and obeying instructions in `AGENTS.md` and `specs/README.md`. In practice, agents often forget to check off tasks in spec files and commit the spec update alongside their implementation changes.

Rather than relying on documentation compliance or moving workflow knowledge into CLI help output, we should ship Specture's workflow as an **agent skill** — a standardized format ([agentskills.io](https://agentskills.io)) that agents load automatically into their context.

The skill replaces the need for:

- Detailed `AGENTS.md` / `CLAUDE.md` workflow prompts
- Comprehensive `specs/README.md` in every project
- A `done` CLI command for checking off tasks
- A pre-commit hook for nudging agents

Instead, agents get the full Specture workflow injected into their context at the right time, and they follow it natively.

## Design Decisions

### Agent skill instead of CLI help output

- Chosen: Ship a `specture` agent skill
  - Skills are loaded into agent context automatically — zero discovery friction
  - Progressive disclosure: agents see only the name/description at startup, full instructions load on activation
  - The skill is always up-to-date (installed version = current instructions), eliminating doc drift
  - Works with the agent's natural capabilities — editing files, running commands — rather than wrapping them in CLI subcommands
  - Open standard with growing adoption across agent platforms
- Considered: Move workflow knowledge into `specture help` output
  - Requires agents to discover and run the command
  - Help output is optimized for humans, not agent context
  - Agents must parse CLI output rather than receiving structured instructions
- Considered: Keep detailed `AGENTS.md` prompt and comprehensive `specs/README.md`
  - Relies on agents reading and following docs
  - Prompt goes stale when Specture evolves
  - Different projects may have outdated versions of the prompt

### No `done` command

- Chosen: Skill instructs agents to edit spec files directly
  - Agents are already excellent at editing markdown files — this is a core capability
  - No new CLI command to discover, learn, or maintain
  - The skill provides clear instructions: change `- [ ]` to `- [x]`, stage the file, commit with the implementation
  - Fewer moving parts = less to break
- Considered: `specture done [substring]` CLI command
  - Adds complexity for something agents can already do
  - Requires substring matching logic, ambiguity handling, error messages
  - The real problem was agents not knowing the workflow, not lacking a tool to execute it

### No pre-commit hook for workflow nudging

- Chosen: Rely on skill instructions instead of hook-based reminders
  - Skills inject workflow knowledge before the agent starts working, not after
  - Pre-commit hooks are reactive (remind after forgetting); skills are proactive (prevent forgetting)
  - No hook configuration or framework integration needed
  - The existing `validate` pre-commit hook remains for format validation — that's a CI concern, not a workflow concern
- Considered: Non-blocking pre-commit hook (prints reminder, exits 0)
  - Only helps if agents read commit output carefully
  - Adds installation complexity
  - Agents may still forget if the reminder isn't actionable in context

### Skill installation via `specture setup`

- Chosen: `specture setup` installs the skill into the project's `.skills/` directory
  - Follows the agent skills convention for project-local skills
  - Skill is versioned with the project (committed to git)
  - `specture setup` already creates `specs/` directory and templates — adding the skill is a natural extension
  - Updating Specture and re-running setup updates the skill
- Considered: Install to `~/.skills/` (user-global)
  - Not project-specific — can't version with the project
  - Assumes a specific agent's directory conventions
- Considered: Skill lives only in the Specture repo, agents fetch it remotely
  - Adds network dependency
  - Projects can't customize or pin a version

### Simplified project docs

- Chosen: Minimal `AGENTS.md` / `CLAUDE.md` that just mentions Specture and the skill
  - One or two lines pointing agents to the skill
  - Project-specific development instructions stay in `AGENTS.md` as before
  - `specs/README.md` becomes a brief overview for humans, linking to the Specture repo
  - All workflow detail lives in the skill where agents actually consume it
- Considered: Remove `AGENTS.md` content entirely
  - Projects still need project-specific instructions (build commands, test setup, etc.)
  - The Specture section just gets much smaller

### Skill structure

The skill follows the [Agent Skills specification](https://agentskills.io/specification.md):

```
.skills/specture/
├── SKILL.md              # Workflow instructions for agents
└── references/
    └── spec-format.md    # Spec file format reference (loaded on demand)
```

- `SKILL.md` contains the core workflow: how to implement specs, check off tasks, commit properly, use CLI commands
- `references/spec-format.md` contains the detailed spec file format (frontmatter fields, sections, naming conventions) — loaded only when creating or editing specs
- Keeps `SKILL.md` focused and under 500 lines per the spec recommendation

## Task List

### Skill Content

- [ ] Write `SKILL.md` with frontmatter (name: `specture`, description: "Follow the Specture System for spec-driven development. Use when creating, implementing, or managing specs.")
- [ ] Write core workflow instructions: implementing specs, checking off tasks, committing properly
- [ ] Document CLI commands in skill (`specture status`, `specture new`, `specture validate`)
- [ ] Write `references/spec-format.md` with detailed spec file format (frontmatter, sections, naming, precedence)
- [ ] Validate skill against the Agent Skills specification

### Skill Installation

- [ ] Embed skill files in Go binary and implement `InstallSkill` (writes `.skills/specture/SKILL.md`)
- [ ] Write reference files (`.skills/specture/references/spec-format.md`)
- [ ] Overwrite existing skill files on re-run
- [ ] Support dry-run flag
- [ ] Integrate `InstallSkill` into `specture setup` flow

### Simplify Project Docs

- [ ] Reduce `agent-prompt.md` template to a minimal Specture mention
- [ ] Slim down `specs-readme.md` template to a brief overview linking to the Specture repo
- [ ] Update `specture setup` to generate the simplified docs
