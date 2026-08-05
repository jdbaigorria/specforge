# Requirements — {{ .Feature }}
{{ with .Summary }}
{{ . }}
{{ end }}{{ if .Actors }}
**Actors:** {{ list .Actors }}
{{ end }}{{ range .Requirements }}
## {{ .ID }} ({{ .EarsType }})

{{ ears . }}
{{ if .Acceptance }}
**Acceptance:**
{{ range .Acceptance }}- {{ with .ID }}**{{ . }}** — {{ end }}{{ criterion . }}
{{ end }}{{ end -}}
{{ end }}
