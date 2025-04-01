package information

import (
	"bytes"
	"html/template"
	"log"
	"reflect"

	"github.com/Masterminds/sprig/v3"
)

const defaultTemplate = `
{{.Title}}
{{ .Text }}
{{- range $it := .Items }}
* {{ $it }}
{{- end }}
{{ .RenderSubsections }}
`

type Information struct {
	OS []string
	Section
}

type Section struct {
	Title        string     `yaml:"Title,omitempty"`
	SubSections  []*Section `yaml:"SubSections,omitempty"`
	EmbeddingVec []float64  `yaml:"EmbeddingVec,omitempty"`
	Text         string     `yaml:"Text,omitempty"`
	Items        []string   `yaml:"Items,omitempty"`
	Files        []string   `yaml:"Files,omitempty"`
	Commands     []string   `yaml:"Commands,omitempty"`
}

func (info *Section) Render(args ...any) string {
	tmpl := defaultTemplate
	for _, arg := range args {
		switch t := arg.(type) {
		case string:
			tmpl = t
		}
	}
	funcMap := template.FuncMap{
		"RenderSubsections": info.RenderSubsections,
	}
	for key, value := range sprig.TxtFuncMap() {
		funcMap[key] = value
	}
	template, err := template.New("sections").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		log.Printf("couldn't parse template: %s\n", err)
	}
	var buf bytes.Buffer
	if err := template.Execute(&buf, info); err != nil {
		log.Printf("couldn't render template: %s\n", err)
	}
	return buf.String()
}

func (info *Section) RenderSubsections() (ret string) {
	for _, sec := range info.SubSections {
		ret += sec.Render()
	}
	return
}

// func (info *Section) ToText() (text string) {
// 	if info.Title != "" {
// 		text += info.Title
// 	}
// 	text += info.Text
// 	for _, sec := range info.SubSections {
// 		text += sec.ToText()
// 	}
// 	for _, it := range info.Items {
// 		text += "* " + it
// 	}
// 	return
// }

func Flatten(info any) {
	typ := reflect.TypeOf(info)
	val := reflect.ValueOf(info)
	for i := 0; i < val.NumField(); i++ {
		if typ.Field(i).Type.Kind() == reflect.Array {
			if val.Len() == 0 {
				val.Index(i).Set(reflect.Zero(typ.Field(i).Type))
			} else {
				Flatten(val.Index(i).Interface)
			}
		}
	}
}
