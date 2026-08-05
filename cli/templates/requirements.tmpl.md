# Requirements — {{ .Feature }}
{{ with .Summary }}
{{ . }}
{{ end }}{{ if .Actors }}
**Actors:** {{ list .Actors }}
{{ end }}{{ range .Requirements }}
## {{ .ID }} ({{ .EarsType }}{{ with .Priority }} · {{ . }}{{ end }})

{{ ears . }}
{{ if .Acceptance }}
**Acceptance:**
{{ range .Acceptance }}- {{ with .ID }}**{{ . }}** — {{ end }}{{ criterion . }}
{{ end }}{{ end -}}
{{ end }}
