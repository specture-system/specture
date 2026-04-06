# Spec Hierarchy Progress

This file tracks the remaining transition work for spec 8.

Current state:

- Recursive discovery is in place.
- `status`, `validate`, and `implement` can resolve dotted spec references.
- `list` still shows top-level specs only.
- `shouldIncludeSpecPath` still accepts root-level flat spec markdown files as a temporary compatibility layer.

Next cleanup:

- Remove the root-level flat markdown allowance from `shouldIncludeSpecPath` once the repo is fully migrated to nested `SPEC.md` files.
- Narrow discovery to the hierarchy-only layout after the migration is complete.
