# Migration Workflow

Use this workflow when bootstrapping Specture in an existing repository or normalizing an older specs layout.

## Bootstrap a Specs Tree

1. Create `specs/` if it does not exist.
2. Add `specs/README.md` explaining the local spec process.
3. Add `specs/.gitignore` using [specs-gitignore-format.md](specs-gitignore-format.md).
4. Create specs as directories containing `SPEC.md` files.
5. Run `specture validate`.

## Migrate Flat Spec Files

For old files such as `specs/003-status-command.md`:

1. Create a directory using the same ref and slug, such as `specs/003-status-command/`.
2. Move the file to `specs/003-status-command/SPEC.md`.
3. Remove legacy `number` frontmatter; refs are derived from directory names.
4. Update markdown links to repo-root-relative `SPEC.md` paths, such as `specs/003-status-command/SPEC.md`.
5. Run `specture validate`.
