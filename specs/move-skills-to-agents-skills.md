---
status: completed
author: Addison Emig
creation_date: 2026-02-11
---

# Move Skills to .agents/skills/

The Specture skill is currently installed to `.skills/specture/`. The emerging convention for agent configuration uses `.agents/` as the top-level directory, with skills nested under `.agents/skills/`. This is the layout used by [agentskills.io](https://agentskills.io) and increasingly adopted by agent platforms.

We should move the skill installation target from `.skills/` to `.agents/skills/` to align with this convention. This also positions the `.agents/` directory as the natural home for future agent configuration (e.g., tool definitions, agent profiles) without cluttering the project root with multiple dotfiles.

## Design Decisions

### Installation target directory

- Chosen: `.agents/skills/`
  - Matches the common `.agents/` convention for agent configuration
  - Skills are clearly namespaced under a broader agent config directory
  - Room for future `.agents/` contents (tool configs, profiles, etc.) without new top-level dotdirs
- Considered: Keep `.skills/`
  - Simpler, already implemented
  - Less conventional — most agent tooling is converging on `.agents/`

### Migration of existing `.skills/` directories

- Chosen: Automatically migrate during `specture setup`
  - If `.skills/specture/` exists and `.agents/skills/specture/` does not, move the files and remove the old directory
  - Print a message explaining the migration
  - Skip migration if both exist (user may have customized)
  - Respect dry-run flag
- Considered: No automatic migration
  - Users would have stale `.skills/` directories after updating
  - Manual cleanup is easy to forget
- Considered: Warn but don't migrate
  - Less disruptive but leaves the old directory around

### Clean up empty `.skills/` directory after migration

- Chosen: Remove `.skills/` if it's empty after migration
  - Clean project root
  - Only remove if empty — don't delete user files
- Considered: Always leave `.skills/` in place
  - Safer but leaves an empty directory that serves no purpose

## Task List

### Core Changes

- [x] Update `InstallSkill` in `internal/setup/skill.go` to write to `.agents/skills/` instead of `.skills/`
- [x] Update `InstallSkill` tests to expect files under `.agents/skills/`

### Migration

- [x] Add `MigrateSkillsDir` function that moves `.skills/specture/` to `.agents/skills/specture/` if the old path exists and the new one doesn't
- [x] Remove `.skills/` directory if empty after migration
- [x] Support dry-run flag in migration
- [x] Write tests for migration (old exists, new doesn't; both exist; old doesn't exist; non-empty `.skills/` after move)
- [x] Call `MigrateSkillsDir` from `specture setup` before `InstallSkill`

### Update References

- [x] Update `agent-prompt.md` template to reference `.agents/skills/specture/` instead of `.skills/specture/`
- [x] Update `AGENTS.md` and `CLAUDE.md` in this repo to reference the new path
