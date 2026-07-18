# Split Workflow

Use this workflow when the user asks to split an existing spec into parent and child specs. Follow [design-workflow.md](design-workflow.md) for general design, discussion, approval, and validation behavior; this workflow defines how to redistribute established content without requiring the user to discuss it again.

## Steps

1. Read the complete source `SPEC.md` and treat its existing goals, requirements, alternatives, and decisions as established design context.
2. Discuss the proposed child specs and their boundaries with the user. Do not infer boundaries from the source spec, and do not create children until the user explicitly confirms the split.
3. Create each child with `specture new --title "Child feature" --parent <source-ref>`.
4. Populate each child with the relevant established content from the source spec. Preserve its meaning; do not invent, expand, or redesign it.
5. Convert the source into a parent spec using the parent structure in [spec-format.md](spec-format.md):
   - simple description
   - `## Goals`
   - `## Child Specs` with title links to every direct child
6. Remove design decisions from the parent once they are represented in the appropriate child specs.
7. Discuss any newly exposed design gaps interactively and record new decisions only after the user confirms them.
8. Run `specture validate` after editing the spec tree.

Explicitly confirm the child boundaries, but do not ask the user to reconfirm established design content merely because it moved from the source spec into a child. Splitting authorizes structural redistribution, not changes to the established design.
