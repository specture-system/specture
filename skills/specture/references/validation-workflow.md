# Validation Workflow

Use validation after editing `SPEC.md`, `PLAN.md`, or migrating the specs tree.

## Commands

```bash
specture validate
specture validate --spec 11
```

Use the narrowest validation that covers the edited files. For broad migrations, validate the whole specs tree.

## What Validation Proves

Validation checks the structural rules Specture can enforce, including parseable frontmatter, valid statuses, required descriptions, duplicate references, and supported spec tree layout.

Validation does not prove implementation correctness. Pair it with project tests, type checks, or linters when code changed.

## After Failures

Read the reported path and field, fix the source file, then rerun the same validation command. Do not paper over validation errors by removing useful spec content.
