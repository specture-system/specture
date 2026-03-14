Review the repository's current working tree for one Specture task.

Spec Path: {{.SpecPath}}
Section: {{.SectionName}}
Task: {{.TaskText}}

Review rules:
- Evaluate whether the current repository state correctly and completely satisfies this task.
- Treat previously accepted changes in the branch as valid baseline context unless this task introduced a regression in them.
- Ignore unrelated pre-existing changes unless they prevent this task from being correct or releasable.
- Only block on critical issues: task not fulfilled, correctness problems, security issues, data loss risks, or build/test breakage caused by this task.
- Ignore nits, style preferences, and non-critical follow-up suggestions.

Respond with:
- REVIEW_CRITICAL: <brief reason>   if critical issues remain
- REVIEW_OK                         if the task is acceptable
