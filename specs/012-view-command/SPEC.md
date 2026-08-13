---
status: completed
author: Addison Emig
creation_date: 2026-07-18
approved_by: Addison Emig
approval_date: 2026-07-18
---

# View Command

Users who discover a spec through `specture list` should be able to open it without
manually copying or reconstructing its path. Add a `view` subcommand that resolves a
spec reference and opens that spec's `SPEC.md` in the user's configured editor.

For example, `specture view 4.2` opens the `SPEC.md` belonging to spec `4.2` in
the user's preferred interactive editor.

## Goals

- Let users select a top-level or nested spec by reference, such as `4.2`.
- Open the selected spec's `SPEC.md` in the editor configured by `$VISUAL`, falling
  back to `$EDITOR`, or use `cat` when neither variable is set.

## Design Decisions

### Accept spec references only

- Chosen: Accept exactly one spec reference, such as `4` or `4.2`.
  - Resolving a spec reference to its file is the value provided by this command.
- Considered: Also accept a directory or file path.
  - Users who already have a path can open it directly and do not need Specture to
    resolve it.

### Prefer VISUAL over EDITOR

- Chosen: Use `$VISUAL` when set and fall back to `$EDITOR`.
  - The command opens an interactive editor, so `$VISUAL` is the more specific Unix
    convention while `$EDITOR` remains a widely supported fallback.
- Considered: Use only `$EDITOR`.
  - This is simpler but ignores users who intentionally configure a different
    full-screen interactive editor.

### Support arguments in the editor command

- Chosen: Treat the selected editor variable as a command that may include arguments,
  then append the resolved `SPEC.md` path.
  - This supports common configurations such as `VISUAL="code --wait"` and
    `EDITOR="nvim -f"`.
- Considered: Treat the full variable value as an executable path.
  - This would mistake commonly used editor flags for part of the executable name.

### Honor the editor command's lifetime

- Chosen: Attach the editor command to standard input, output, and error, and wait for
  the command to exit.
  - Terminal editors remain interactive, while graphical editor launchers can return
    immediately or wait according to their own arguments.
- Considered: Launch the editor as a detached process and return immediately.
  - Detaching would prevent terminal editors from using the terminal correctly and
    override the behavior selected by the user's editor command.

### Default to cat

- Chosen: Use `cat` when neither `$VISUAL` nor `$EDITOR` is set.
  - This makes `specture view` useful in non-interactive environments and does not
    require an editor to be installed.
- Considered: Return an actionable error when neither variable is set.
  - Requiring configuration prevents users from quickly viewing a spec in a
    terminal or script.
- Considered: Fall back to a default editor such as `vi`.
  - An implicit editor could be surprising and may not exist in the user's
    environment.
