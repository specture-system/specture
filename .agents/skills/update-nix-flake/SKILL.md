---
name: update-nix-flake
description: Use when modifying flake.nix, go.mod, or Go dependencies in the Specture repository, especially when vendorHash must be refreshed for Nix builds.
---

# Update Nix Flake

This skill covers Specture's repo-specific Nix packaging workflow.

Use it when:

- `flake.nix` changes
- `go.mod` or `go.sum` changes
- Nix builds fail because `vendorHash` is stale
- release or packaging work touches the flake

## Rules

- Keep `flake.nix` and Go module metadata in sync.
- Do not run `go mod vendor`. This repository does not keep a vendored tree.
- Prefer the repo's normal development workflow for everything else. This skill is only for the Nix-specific parts.

## Updating `vendorHash`

When Go dependencies change, refresh `vendorHash` in [`flake.nix`](../../../flake.nix):

1. Set `vendorHash = "";`.
2. Run `nix build`.
3. Read the error output and copy the `got:` hash.
4. Replace the empty `vendorHash` with that hash.
5. Run `nix build` again to verify the flake succeeds.

Expected error shape:

```text
specified: sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
got: sha256-ZknVM8bMM0kLIbuV4Bv4XsbgtyhlKyP7p2AVOE1k0GA=
```

Use the `got:` value.

## Validation Checklist

- `flake.nix` still evaluates
- `nix build` succeeds
- `vendorHash` matches the current module graph
- no `vendor/` directory was added

## Notes

- `flake.nix` currently builds the CLI with `buildGoModule`.
- The flake also defines the default package, app, and dev shell, so keep those intact unless the task requires changing them.
