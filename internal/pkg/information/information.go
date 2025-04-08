package information

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"html/template"
	"log"
	"reflect"
	"strings"

	"github.com/Masterminds/sprig/v3"
	"github.com/mslacken/kowalski/internal/app/ollamaconnector"
)

const defaultTemplate = `
{{- if .Title }}{{ Section }} {{.Title}}{{ end }}
{{- if .Text }}{{ .Text }}{{ end}}
{{- range $it := .Items }}
* {{ $it }}
{{- end }}
{{ RenderSubsections .Level }}
`

type RenderData struct {
	Level int
	Section
}

type Information struct {
	OS   []string
	Hash string
	Dist float32
	Section
}

type Section struct {
	Title        string     `yaml:"Title,omitempty"`
	SubSections  []*Section `yaml:"SubSections,omitempty"`
	EmbeddingVec []float32  `yaml:"EmbeddingVec,omitempty"`
	Text         string     `yaml:"Text,omitempty"`
	Items        []string   `yaml:"Items,omitempty"`
	Files        []string   `yaml:"Files,omitempty"`
	Commands     []string   `yaml:"Commands,omitempty"`
}

func (info *Section) Render(args ...any) string {
	level := 0
	tmpl := defaultTemplate
	for _, arg := range args {
		switch t := arg.(type) {
		case string:
			tmpl = t
		case int:
			level = t
		}
	}
	funcMap := template.FuncMap{
		"RenderSubsections": info.RenderSubsections,
		"Section": func() string {
			return strings.Repeat("#", level)
		},
	}
	for key, value := range sprig.TxtFuncMap() {
		funcMap[key] = value
	}
	template, err := template.New("sections").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		log.Printf("couldn't parse template: %s\n", err)
	}
	var buf bytes.Buffer
	if err := template.Execute(&buf, RenderData{
		Section: *info,
		Level:   level,
	}); err != nil {
		log.Printf("couldn't render template: %s\n", err)
	}
	return strings.Replace(buf.String(), "\n\n", "\n", -1)
}

func (info *Information) Empty() bool {
	return len(info.SubSections) == 0 &&
		info.Text == ""
}

func (info *Section) RenderSubsections(level int) (ret string) {
	for _, sec := range info.SubSections {
		ret += sec.Render(level + 1)
	}
	return
}

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

func (info *Information) CreateHash() []byte {
	str := info.Render()
	h := sha256.New()
	h.Write([]byte(str))
	info.Hash = fmt.Sprintf("%x", h.Sum(nil))
	return h.Sum(nil)
}

func (info *Information) CreateEmbedding() (emb []float32, err error) {
	str := info.Render()
	embResp, err := ollamaconnector.Ollama().GetEmbeddings([]string{str})
	if err != nil {
		return nil, err
	}
	if len(embResp.Embeddings) == 0 {
		return nil, fmt.Errorf("couldn't calculate embedding")
	}
	info.EmbeddingVec = embResp.Embeddings[0]
	return embResp.Embeddings[0], nil
}
