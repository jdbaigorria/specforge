# Review — {{ .Feature }}

**Verdict:** {{ .Verdict }}
{{ with .VerdictRationaleMD }}
{{ . }}
{{ end }}{{ if .Traceability }}
## Traceability
{{ range .Traceability }}- {{ .Requirement }} → tasks: {{ list .Tasks }} · files: {{ list .Files }} · tests: {{ list .Tests }} · **{{ .Status }}**
{{ end }}{{ end }}{{ if .Gaps }}
## Gaps
{{ range .Gaps }}- {{ . }}
{{ end }}{{ end }}{{ if .ConstitutionViolations }}
## Constitution violations
{{ range .ConstitutionViolations }}- {{ . }}
{{ end }}{{ end -}}
