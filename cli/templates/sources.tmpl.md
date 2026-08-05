# Sources

Material this project's requirements were derived from. Each requirement cites
the ids it came from — a requirement citing none was invented by the model.

| ID | Kind | Ref | Captured | Note |
|---|---|---|---|---|
{{ range .Sources }}| {{ .ID }} | {{ .Kind }} | `{{ .Ref }}` | {{ orDash .Captured }} | {{ orDash .Note }} |
{{ end }}
