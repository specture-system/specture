# Implementation Workflow

Use this workflow when implementing an `approved` or `in-progress` spec. If a sibling `PLAN.md` exists, treat it as execution context for that spec, not as a separate implementation authority.

## Starting Implementation

1. Read the relevant `SPEC.md`.
2. Update its frontmatter `status` to `in-progress`.
3. Create or select an implementation branch using the project's branch naming conventions.

## Execution Loop

Follow this loop for every implementation chunk. Do not skip the commit step.

1. Read the spec and any sibling `PLAN.md`.
2. Select exactly one small implementation chunk.
3. Analyze only enough code to implement that chunk safely.
4. Implement the chunk.
5. Update tests, docs, or `PLAN.md` when needed.
6. Run the narrowest verification that proves the chunk.
7. Commit the focused change before starting another chunk.
8. Repeat from step 2 until the spec goals are complete.

If a chunk becomes too large or mixes unrelated concerns, stop and split it before committing.

## Pull Request Plans

When a spec's `PLAN.md` divides work into PRs or chunks, treat each bullet group as a commit boundary unless the plan says otherwise. A working tree should normally contain only the current focused chunk.

## Completing Implementation

Only mark the spec `completed` after all planned behavior is implemented and validated. Keep the completion update separate when that makes review easier.
