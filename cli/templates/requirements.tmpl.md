# Requirements — {{ .Feature }}
{{ with .Summary }}
{{ . }}
{{ end }}{{ if .Actors }}
**Actors:** {{ list .Actors }}
{{ end }}{{ range .Requirements }}
## {{ .ID }} ({{ .EarsType }}{{ with .Priority }} · {{ . }}{{ end }}{{ with .Kind }} · {{ . }}{{ end }}{{ with .Verification }} · verified by {{ . }}{{ end }})

{{ ears . }}
{{ with source . }}
**Source:** {{ . }}
{{ end }}{{ if .Acceptance }}
**Acceptance:**
{{ range .Acceptance }}- {{ with .ID }}**{{ . }}** — {{ end }}{{ criterion . }}
{{ end }}{{ end -}}
{{ end }}
