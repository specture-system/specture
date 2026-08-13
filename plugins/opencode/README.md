# Specture Agents for OpenCode

Provides three OpenCode agents — `specture-design`, `specture-implement`, and `specture-review` — that encode the Specture spec-driven development workflow.

## Install

```sh
# Project-level
mkdir -p .opencode/agents
curl -fsSL https://raw.githubusercontent.com/specture-system/specture/main/plugins/opencode/agents/specture-design.md -o .opencode/agents/specture-design.md
curl -fsSL https://raw.githubusercontent.com/specture-system/specture/main/plugins/opencode/agents/specture-implement.md -o .opencode/agents/specture-implement.md
curl -fsSL https://raw.githubusercontent.com/specture-system/specture/main/plugins/opencode/agents/specture-review.md -o .opencode/agents/specture-review.md

# Or global
mkdir -p ~/.config/opencode/agents
curl -fsSL https://raw.githubusercontent.com/specture-system/specture/main/plugins/opencode/agents/specture-design.md -o ~/.config/opencode/agents/specture-design.md
curl -fsSL https://raw.githubusercontent.com/specture-system/specture/main/plugins/opencode/agents/specture-implement.md -o ~/.config/opencode/agents/specture-implement.md
curl -fsSL https://raw.githubusercontent.com/specture-system/specture/main/plugins/opencode/agents/specture-review.md -o ~/.config/opencode/agents/specture-review.md
```

OpenCode loads agent markdown files from these directories automatically at startup. The filename (minus `.md`) becomes the agent name.
