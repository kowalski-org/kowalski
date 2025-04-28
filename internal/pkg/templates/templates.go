package templates

const RenderInfo = `
# {{ .Title }} {{ range $it := .Lines }}
{{ if eq $it.Type "command" }}'''{{ $it.Text}}'''{{ else }}{{ $it.Text}}{{ end }}
{{- end }}`

const RenderInfoWithMeta = RenderInfo + `
{{- if .Files}}
Files:{{ else }}
No files{{ end }}
{{- range $file := .Files}}
* {{ $file }}
{{- end }}
{{- if .Commands }}
Commands:{{ else }}
No commands{{ end }}
{{- range $cmd := .Commands}}
* {{ $cmd }}
{{- end }}
`

const RenderTitleOnly = `
Source: {{ .Source }}
{{ if .OS }}OS: {{ range $os := .OS}}{{ $os }}{{ end }}{{ end }}
{{ range $sec := .Sections }}
{{ $sec.Title }}
{{ end }}
`

const Prompt = `Your name is Kowlaski and you are a helpfull assistant for a {{ .Name }} {{ .Version }} system.
Answer in short sentences.
If your answer contains a shell command start it with <command> and end it with </command>.
If you answer contains a new configuration start the changed file with <file id=filename> and end it with </file>.
{{ .Context }}
The user wants help with following task:
{{ .Task }}`
