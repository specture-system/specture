Run one final cleanup worker pass after all Specture sections are complete.

Spec Path: {{.SpecPath}}
Current branch: {{.CurrentBranch}}
Completed sections:
{{range .Sections}}- {{.}}
{{end}}

Cleanup review recommendations:
{{.ReviewOutput}}

Instructions:
- Treat the current checkout as the source of truth. Preserve existing accepted section work already present in the branch.
- Implement only the recommended cleanup refactors.
- Keep cleanup changes low risk and maintainability-focused.
- Do not edit the spec file.
- Do not create commits.
- Do not run additional review loops.
- Stop after this single cleanup worker pass and leave changes in the working tree.
