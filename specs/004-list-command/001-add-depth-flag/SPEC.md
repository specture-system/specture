---
status: draft
author: Addison Emig
creation_date: 2026-06-18
---

# Add Depth Flag

Add a `-d`/`--depth` flag to the list command that controls how deep into the spec hierarchy to recurse. Default value is 1, which preserves the current behavior (top-level specs only, or immediate children when `--parent` is set).

## Design Decisions

- **`-d`/`--depth` defaults to 1.** This matches the current behavior where `list` shows only specs at the current level without recursing into children.
  - Exception: if `--parent` is passed, `--depth` defaults to `all`. `--parent` will already filter the output significantly so we can use `--depth all` by default without overwhelming the user.

- **`--depth all` means unlimited.** Most readable option for viewing all specs.
- **`--depth 0` is accepted as an alias for `all`.** Zero would otherwise mean "show nothing", so it maps to unlimited as a convenience.
