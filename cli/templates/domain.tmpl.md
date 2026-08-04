# Domain knowledge
{{ if .Glossary }}
## Glossary
{{ range .Glossary }}- **{{ .Term }}** — {{ .Definition }}{{ if .Aliases }} _(aka: {{ list .Aliases }})_{{ end }}
{{ end }}{{ end }}{{ if .Entities }}
## Entities
{{ range .Entities }}- **{{ .ID }} {{ .Name }}** — {{ .Description }}{{ range .Invariants }}
  - _invariant:_ {{ . }}{{ end }}
{{ end }}{{ end }}{{ if .Rules }}
## Business rules
{{ range .Rules }}- **{{ .ID }}** {{ .Rule }}{{ if .Entities }} _(entities: {{ list .Entities }})_{{ end }}{{ if .AppliesTo }} _(audits: {{ list .AppliesTo }})_{{ end }}
{{ end }}{{ end -}}
