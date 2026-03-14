Run one final cleanup review across all completed Specture sections.

Spec Path: {{.SpecPath}}
Current branch: {{.CurrentBranch}}
Completed sections:
{{range .Sections}}- {{.}}
{{end}}

Cleanup review focus:
- Identify unnecessary abstraction that can be simplified safely.
- Identify clear AGENTS.md guideline violations in the completed implementation.
- Identify low-risk maintainability improvements that do not change intended behavior.

Instructions:
- Treat the current checkout as the source of truth.
- This is a bounded final cleanup review pass, not another acceptance gate.
- Suggest only concrete, low-risk refactors that can be done in one follow-up worker pass.

Respond with concise actionable recommendations for the cleanup worker pass.
