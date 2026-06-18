# Tasks — {{ .Feature }}
{{ range .Waves }}
## Wave {{ .N }}{{ with .Name }} — {{ . }}{{ end }}
{{ range .Tasks }}
- **{{ .ID }}** — {{ .Title }} `[{{ .Status }}]`
  - requirements: {{ list .RequirementRefs }}
  - components: {{ list .ComponentRefs }}
  - files: {{ list .FilesTouched }}
  - effort: {{ orDash .EstimatedEffort }}
{{- end }}
{{ end -}}
