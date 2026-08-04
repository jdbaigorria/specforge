# Journal — {{ .Feature }} ({{ .Date }})
{{ range .Lessons }}
## {{ .Rule }}

**Context:** {{ .Context }}
{{ if .Tags }}**Tags:** {{ list .Tags }}
{{ end }}{{ if .Anchors }}**Anchors:** {{ list .Anchors }}
{{ end }}{{ end -}}
