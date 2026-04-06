# Spec Hierarchy Progress

This file tracks the remaining transition work for spec 8.

Current state:

- Recursive discovery is in place and legacy flat-spec support is gone outside `setup` migration.
- `new` supports `--parent` for creating child specs.
- `list` supports `--parent` and shows `ref`, `name`, `status`, and `path` in both text and JSON output.
- `validate` resolves dotted references and validates the whole tree recursively when run without `--spec`.
- `rename` renames spec directories and rewrites repo-root-relative spec links.
- The shipped skill/docs/templates are updated to the new hierarchy wording.

Next cleanup:

- Remove task/progress semantics from the spec model and implementation flow.
  - This branch removes status inference from checkbox completion and stops mutating `SPEC.md` after task acceptance.
  - `internal/spec` still parses task lists for implementation planning.
- After that, remove any remaining flat-layout compatibility scaffolding that is only kept for transition safety.
