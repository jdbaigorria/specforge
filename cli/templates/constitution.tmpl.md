# Constitution
{{ with .IdentityMD }}
{{ . }}
{{ end }}{{ if .Principles }}
## Principles
{{ range .Principles }}- **{{ .ID }}** {{ .Statement }}{{ if .AppliesTo }} _(audits: {{ list .AppliesTo }})_{{ end }}
{{ end }}{{ end }}{{ if .Constraints }}
## Constraints
{{ range .Constraints }}- {{ . }}
{{ end }}{{ end }}{{ if .AntiGoals }}
## Anti-goals
{{ range .AntiGoals }}- {{ . }}
{{ end }}{{ end }}{{ if .Invariants }}
## Invariants
{{ range .Invariants }}- **{{ .ID }}** {{ .Rule }}{{ if .AppliesTo }} _(audits: {{ list .AppliesTo }})_{{ end }}{{ with .At }} ({{ . }}){{ end }}
{{ end }}{{ end -}}
