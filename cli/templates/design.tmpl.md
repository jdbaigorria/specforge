# Design — {{ .Feature }}
{{ with .SummaryMD }}
{{ . }}
{{ end }}{{ if and (section .SectionsIncluded "components") .Components }}
## Components
{{ range .Components }}
### {{ .ID }} — {{ .Name }}{{ with .Kind }} ({{ . }}){{ end }}
{{ range .Responsibilities }}- {{ . }}
{{ end }}{{ if .DependsOn }}depends on: {{ list .DependsOn }}
{{ end }}{{ end }}{{ end }}{{ if and (section .SectionsIncluded "decisions") .Decisions }}
## Decisions
{{ range .Decisions }}
### {{ .ID }} — {{ .Question }}
{{ range .Options }}- **{{ .Name }}**{{ if .Pros }} (pros: {{ list .Pros }}){{ end }}{{ if .Cons }} (cons: {{ list .Cons }}){{ end }}
{{ end }}**Chosen:** {{ .Chosen }}{{ with .RationaleMD }} — {{ . }}{{ end }}
{{ end }}{{ end }}{{ if and (section .SectionsIncluded "interfaces") .Interfaces }}
## Interfaces
{{ range .Interfaces }}- `{{ .Name }}`
{{ end }}{{ end -}}
