---
status: completed
---

# Assignee Support

Some projects already use an `assignee` frontmatter field to record who owns a spec, but Specture does not currently expose that metadata when querying the spec tree. This makes it impossible to use the CLI to find the specs assigned to a particular person.

## Goals

- Treat `assignee` as supported optional spec metadata.
- Allow specs to be filtered by assignee from the CLI.
- Display assignments in text list output without adding an empty column when assignees are not in use.

## Design Decisions

1. Add `--assignee` to `specture list` and its existing `ls` alias. It accepts one or more comma-separated names. Match when a spec's assignee case-insensitively equals any complete trimmed filter value. Do exact matches, not partial-name matches. This uses the existing `--status` comma-separated convention, but unlike the current status implementation assignee matching is explicitly case-insensitive.
2. Add an `ASSIGNEE` column to text list output only when at least one spec in the final displayed result set has an assignee. Specs excluded by status, parent, depth, or assignee filters must not affect column presence.
3. Always include an `assignee` string in every `specture list --format json` entry, using `""` for an unassigned spec.
4. Do not add filtering for unassigned specs (no reserved value and no separate flag).

## Implementation Expectations Based on Current Architecture

- Extend the list-facing spec frontmatter parser/model in `internal/spec` to expose assignee.
- Integrate filtering into `cmd/list.go` after scope/depth/status selection and before either renderer.
- Update CLI long help/examples and supported optional-frontmatter documentation (`skills/specture/references/spec-format.md`). Do not add a blank assignee to the new-spec template unless the design or existing conventions clearly require it.
- Add focused parser and list tests covering exact case-insensitive matching, comma-separated/trimmed filters, no partial matches, conditional text column after all filters, and stable JSON output for assigned and unassigned specs. Preserve all existing output when no displayed spec has an assignee, except for the intentionally expanded JSON schema.
- Use `just` recipes, never raw `go` commands. Run `just run validate --spec 13` after spec edits and the narrowest relevant test recipe(s), then broader validation if warranted.
