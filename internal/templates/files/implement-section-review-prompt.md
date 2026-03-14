Review the repository's current branch state for one completed Specture section.

Spec Path: {{.SpecPath}}
Section: {{.SectionName}}
Tasks:
{{range .Tasks}}- {{.}}
{{end}}

Review rules:
- Evaluate whether the current repository state correctly and completely satisfies the full section.
- Treat previously accepted task commits in this branch as valid baseline context unless the section as a whole is still incorrect.
- Only block on critical issues: incomplete section behavior, correctness problems across tasks, security issues, data loss risks, or build/test breakage caused by this section.
- Ignore nits, style preferences, and non-critical follow-up suggestions.

Respond with:
- REVIEW_CRITICAL: <brief reason>   if critical issues remain
- REVIEW_OK                         if the section is acceptable
