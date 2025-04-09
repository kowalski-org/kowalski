package templates

const RenderInfo = `
{{- if .Title }}{{ Section }} {{.Title}}{{ end }}
{{- if .Text }}{{ .Text }}{{ end}}
{{- range $it := .Items }}
* {{ $it }}
{{- end }}
{{ RenderSubsections .Level }}
`

const SystemPrompt = `You are an helpfull assistant for linux systems. Answer in short sentences.
`
