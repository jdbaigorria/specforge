# Plan — {{ .Feature }}
{{ range .Waves }}
## Wave {{ .N }}{{ with .Name }} — {{ . }}{{ end }}{{ with .Complexity }} (complexity: {{ . }}){{ end }}
- tasks: {{ list .Tasks }}
{{ with .Rationale }}{{ . }}
{{ end }}{{ end -}}
