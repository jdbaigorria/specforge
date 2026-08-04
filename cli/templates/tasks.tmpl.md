# Tasks — {{ .Feature }}
{{ range .Tasks }}
- **{{ .ID }}** — {{ .Title }} `[{{ .Status }}]`
  - requirements: {{ list .RequirementRefs }}
  - components: {{ list .ComponentRefs }}
  - files: {{ list .FilesTouched }}
  - depends on: {{ list .DependsOn }}
  - effort: {{ orDash .EstimatedEffort }}
{{- end }}
