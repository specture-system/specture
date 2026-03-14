Address critical issues found during section-level review for one Specture section.

Spec Path: {{.SpecPath}}
Section: {{.SectionName}}
Tasks:
{{range .Tasks}}- {{.}}
{{end}}

Critical review findings:
{{.ReviewOutput}}

Instructions:
- Treat the current checkout as the source of truth. Preserve existing accepted task commits already present in the branch.
- Make only the changes required to resolve the critical section-level issues above.
- Do not edit the spec file.
- Do not create commits.
- Stop once the critical section-level issues are addressed and leave the fixes in the working tree.
