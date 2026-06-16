---
name: specture
description: Follow the Specture System for spec-driven development. Use when creating, implementing, validating, or migrating specs and agent execution plans.
---

# Specture System

Specture is a spec-driven development system. Durable design records live in `SPEC.md` files under the `specs/` tree. Disposable agent handoffs live in optional `PLAN.md` files beside or beneath specs.

Use the CLI for deterministic file creation, querying, and validation. Use this skill for workflow decisions and file edits.

## Start Here

- Designing or refining a spec: read [references/design-workflow.md](references/design-workflow.md).
- Implementing a spec: read [references/implementation-workflow.md](references/implementation-workflow.md) before changing code, and include any relevant `PLAN.md` as execution context.
- Validating spec or plan edits: read [references/validation-workflow.md](references/validation-workflow.md).
- Bootstrapping or migrating a specs tree: read [references/migration-workflow.md](references/migration-workflow.md).
- Creating or editing a `SPEC.md`: read [references/spec-format.md](references/spec-format.md).
- Creating or editing a `PLAN.md`: read [references/plan-format.md](references/plan-format.md).
- Configuring tracked spec files: read [references/specs-gitignore-format.md](references/specs-gitignore-format.md).

## Core Rules

- Use `specture list` to find specs; do not manually scan the specs tree when the CLI can answer the question.
- Read the relevant `SPEC.md` before implementation work.
- Before implementing any spec, read `references/implementation-workflow.md`; do not treat a `PLAN.md` alone as sufficient workflow guidance.
- During implementation, commit each focused, verified chunk before starting the next chunk.
- Do not batch multiple planned PR chunks into one uncommitted working tree unless the user explicitly asks.
- Keep implementation progress out of `SPEC.md`; use `PLAN.md` for execution handoffs and task breakdowns.
- Do not edit spec design decisions or descriptions without explicit user permission.
- Use plain-language markdown headings; do not number headings.
- Cross-spec mentions must use inline repo-root-relative markdown links to the target `SPEC.md`.
- Run `specture validate` after spec migrations or edits to `SPEC.md`/`PLAN.md` files.

## CLI Quick Reference

```bash
specture list
specture list --status draft,approved
specture list -f json
specture validate
specture validate --spec 11
specture new --title "Feature name"
```

See `specture help` and command-specific `--help` output for exact CLI behavior in the installed version.
