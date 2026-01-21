---
status: draft
author: Addison Emig
creation_date: 2026-01-20
---

# Status Command

Some common questions come up while using Specture to implement a spec:

- What's the current spec?
- What's next for this spec?
- How much progress have we made for this spec?

We should add a `status` command to allow quickly answering these questions without requiring the user to directly read any of the spec files.

## Design Decisions

### Output Format

By default, the output should be in human-readable plain text.

For easy processing by automated tooling we should also support `json` output.

We can use a `--format` flag with values `text` or `json`

### Output Contents

It is useful to include the following:

- Spec name (from top-level header)
- Spec number (from filename)
- Spec status
  - If not specified in frontmatter, we can automatically deduce status using the following simple algorithm:
    - Spec has no task list -> `draft`
    - Spec has no complete tasks -> `draft`
    - Spec has mixture of complete and incomplete tasks -> `in-progress`
    - Spec has only complete tasks -> `completed`
- Current task
  - The contents of first line beneath `## Task List` that starts with `- [ ]`
  - Do not include indented tasks
  - Empty string if no such line is found
- Current task section title
  - Determined by looking for first incomplete task in the task list, then moving up line-by-line to find a section header
  - Empty string if no section header is found between `## Task List` and the incomplete item
- Complete tasks
  - Parse the task list from the markdown and return every one that has been checked off
- Incomplete tasks
  - Parse the task list from the markdown and return every one that has not been checked off

In the future, we may more items to the status output. The current list should be a good start, we aren't trying to be comprehensive with everything that might end up being useful in the status command output in the long run. Future specs can suggest additions to the output.

### Current Spec

By default, this command will return the results for the "current spec".

The current spec is determined by sorting the specs by ascending spec number, then selecting the first that has status `in-progress`.

Status `in-progress` can be inferred even for specs without explicit `status` value in their front matter using the algorithm mentioned above.

A `--spec` flag can be used to get the overall status of any particular spec by number, no matter if it is `in-progress` or not.

## Task List

TBD
