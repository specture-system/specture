# Implementation Workflow

Use this workflow when implementing an `approved` spec.

## Starting Implementation

1. Read the relevant `SPEC.md`.
2. Update its frontmatter `status` to `in-progress`.
3. Create or select an implementation branch using the project's branch naming conventions.
4. Commit the status change separately when the project expects a clean implementation history.

## Execution Loop

1. Read the spec and any sibling `PLAN.md`.
2. Analyze only enough code to identify the next small implementation chunk.
3. Implement the chunk.
4. Update tests, docs, or `PLAN.md` when needed.
5. Run the narrowest verification that proves the chunk.
6. Commit the focused change.
7. Repeat until the spec goals are complete.

## Completing Implementation

Only mark the spec `completed` after all planned behavior is implemented and validated. Keep the completion update separate when that makes review easier.
