Implement exactly one Specture task.

Spec Path: {{.SpecPath}}
Section: {{.SectionName}}
Task: {{.TaskText}}
Task Subtree:
{{.TaskSubtree}}
Current changed files in working tree:
{{- if .ChangedFiles}}
{{range .ChangedFiles}}- {{.}}
{{end}}
{{- else}}
- (none detected)
{{- end}}

{{- if .ReviewOutput}}
Prior critical review findings:
{{.ReviewOutput}}
{{- end}}

Instructions:
- Treat the current checkout as the source of truth. Preserve existing accepted changes already present in the branch.
- Implement only the work required for this task. Do not intentionally modify unrelated behavior or refactor unrelated code.
- If this task requires touching existing code from earlier accepted tasks, make the smallest safe change that builds on that work.
- Do not edit the spec file.
- Do not create commits.
- When the task is amenable to automated testing, follow test-driven development: write a failing test first, implement until it passes, then refactor.
- Stop once this task is fully implemented and the repository is left with the working tree changes for this task.
