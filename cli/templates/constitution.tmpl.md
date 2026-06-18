# Constitution
{{ with .IdentityMD }}
{{ . }}
{{ end }}{{ if .Principles }}
## Principles
{{ range .Principles }}- {{ . }}
{{ end }}{{ end }}{{ if .Constraints }}
## Constraints
{{ range .Constraints }}- {{ . }}
{{ end }}{{ end }}{{ if .AntiGoals }}
## Anti-goals
{{ range .AntiGoals }}- {{ . }}
{{ end }}{{ end }}{{ if .Invariants }}
## Invariants
{{ range .Invariants }}- **{{ .ID }}** {{ .Rule }}{{ with .At }} ({{ . }}){{ end }}
{{ end }}{{ end -}}
