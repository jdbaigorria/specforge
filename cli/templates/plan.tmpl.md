# Plan — {{ .Feature }}
{{ range .Waves }}
## Wave {{ .N }} — complexity: {{ .Complexity }}
{{ with .Rationale }}{{ . }}
{{ end }}{{ end -}}
