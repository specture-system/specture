# Agent-Native Redesign Plan

Implement [Agent-Native Redesign](specs/011-agent-native-redesign/SPEC.md) as small, reviewable commits. Keep each commit focused on one behavior or documentation boundary, and run the narrowest `just` or `specture validate` check that proves the chunk.

## Pull Request Plan

### PR 1: Agent-native skill foundation

- Mark spec 11 as `in-progress`.
- Add this `PLAN.md` execution handoff.
- Update `specs/.gitignore` so `PLAN.md` files are tracked.
- Add `skills/specture/` as the installable Specture skill.
- Add `skills/specture/SKILL.md` as the concise skill entrypoint.
- Add focused references under `skills/specture/references/` for:
  - design workflow
  - implementation workflow
  - validation workflow
  - migration workflow
  - `SPEC.md` format
  - `PLAN.md` format
  - `specs/.gitignore` format
- Update repository guidance and docs to refer to `skills/specture/` as the source of truth.
- Remove legacy `.agents/skills/specture/` after the distributable skill is complete and referenced.
- Verify with the narrowest relevant `just` recipes and `specture validate --spec 11`.

### PR 2: Remove `setup` and simplify `new`

- Remove the `setup` command, setup command registration, setup help, setup implementation, setup tests, setup-only templates, and setup docs.
- Preserve validation and listing behavior.
- Make `specture new` non-interactive and file-only for `SPEC.md` creation.
- Require `--title`.
- Remove title prompt, confirmation prompt, editor launch, stdin body support, dry-run mode, branch creation, `--no-branch`, clean-worktree checks, and branch cleanup.
- Keep auto-allocation and existing parent behavior working.
- Add `specture new --plan` to create `PLAN.md` instead of `SPEC.md`.
- Add the standard plan template.
- Allow `SPEC.md` and `PLAN.md` to coexist in the same numbered directory.
- Fail only when the target file already exists.
- Ensure `PLAN.md` uses spec frontmatter when it stands alone without a sibling `SPEC.md`.
- Add `specture new --spec` / `-s` for explicit refs.
- Support top-level refs such as `123` and dotted child refs such as `123.4`.
- Resolve all parent specs for dotted refs before creating the child.
- Reject `--spec` combined with `--parent`.
- Preserve auto-allocation when `--spec` is omitted.
- Update tests around removed flags and retained/new creation behavior.

### PR 3: Query plans and final cleanup

- Update list and validate behavior for `PLAN.md`.
- Prefer sibling `SPEC.md` when both files exist.
- Query and validate standalone `PLAN.md` only when no sibling `SPEC.md` exists.
- Add tests for standalone plans, coexisting spec/plan files, and invalid plan frontmatter.
- Remove dead helpers, tests, templates, and docs left obsolete by the redesign.
- Run the broadest appropriate `just` verification for the completed change set.
- Mark spec 11 as `completed` only after all planned behavior is implemented and validated.

## Skill Reference Notes

- The `specs/.gitignore` reference should document the expected pattern:

  ```gitignore
  *
  !*/
  !**/SPEC.md
  !**/PLAN.md
  !README.md
  ```
- Explain why this keeps scratch notes and implementation artifacts out of git while preserving durable specs and agent handoff plans.

## Implementation Notes

- Prefer small deletions over compatibility shims when removing unused CLI behavior.
- Keep reusable parsing and discovery logic in `internal/`; keep `cmd/` focused on command orchestration.
- Do not edit spec design decisions unless the implementation exposes a contradiction that needs human approval.
- After changing files under `internal/templates/files/`, run `just run update --yes` so generated repo files stay in sync.
