package information

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/charmbracelet/log"

	"github.com/Masterminds/sprig/v3"
	"github.com/openSUSE/kowalski/internal/app/ollamaconnector"
	"github.com/openSUSE/kowalski/internal/pkg/templates"
)

const maxdirentries = 10
const filemaxsize = 2048

type LineType string

// types for doc book parsing
const (
	// values are gathered from the tags
	Title     LineType = "title"
	Text      LineType = "text"
	Command   LineType = "command"
	File      LineType = "file"
	Formatted LineType = "formatted"
	// this type must be created from it's context
	SubTitle    LineType = "subtitle"
	SubSubTitle LineType = "subsubtitle"
	Warning     LineType = "warning"
)

func (t *LineType) String() string {
	return string(*t)
}

type Line struct {
	Text string
	Type LineType
}

// simplify the text attributes
func GetType(str string) LineType {
	switch strings.ToLower(str) {
	case "title":
		return Title
	case "command", "screen":
		return Command
	case "package", "emphasis", "literal", "option", "replaceable":
		return Formatted
	default:
		return Text
	}
}

type Information struct {
	OS       []string
	Hash     string `boltholdKey:"ID"`
	Source   string
	Sections []Section
	// mentioned files info
	Files []string
	// mentioned commands in info
	Commands []string
}

type Section struct {
	Title        string    `yaml:"Title,omitempty"`
	EmbeddingVec []float32 `yaml:"EmbeddingVec,omitempty"`
	Lines        []Line    `yaml:"Lines,omitempty"`
	Files        []string  `yaml:"Files,omitempty"`
	Commands     []string  `yaml:"Commands,omitempty"`
	IsAlias      bool      `yaml:"IsAlias,omitempty"` // Title is an alias to first section
}

// data returned from db for the LLM modell
type RetSection struct {
	Dist float32
	Hash string // hash which identifies base doc
	Section
}

func (info *Section) Render(args ...any) (string, error) {
	funcMap := sprig.FuncMap()
	tmpl := templates.RenderInfo
	for _, arg := range args {
		switch t := arg.(type) {
		case string:
			tmpl = t
		case map[string]func(string) string:
			for key, val := range t {
				funcMap[key] = val
			}
		}
	}
	template, err := template.New("sections").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		log.Printf("couldn't parse template: %s\n", err)
	}
	var buf bytes.Buffer
	if err := template.Execute(&buf, *info); err != nil {
		return "", err
	}
	return strings.Replace(buf.String(), "\n\n", "\n", -1), nil
}

func (info *Information) Empty() bool {
	return len(info.Sections) == 0
}

func (info *Information) CreateEmbedding(embedding string) (err error) {
	embeddingSize, err := ollamaconnector.Ollamasettings.GetEmbeddingSize(embedding)
	if err != nil {
		return err
	}
	for i, sec := range info.Sections {
		str, err := sec.Render()
		if err != nil {
			return err
		}
		// unfortunately ollama does give the maximal embeddings in tokens
		// and not in chars, so we use just a crude factor between them
		char2TokenMult := uint(3)
		if uint(len(str)) > char2TokenMult*uint(embeddingSize) {
			str = str[:char2TokenMult*embeddingSize]
			log.Warnf("truncated info: %s", sec.Title)
		}
		embResp, err := ollamaconnector.Ollamasettings.GetEmbeddings([]string{str}, embedding)
		if err != nil {
			return err
		}
		if len(embResp.Embeddings) == 0 {
			log.Debugf("embedding text: %s", []string{str})
			return fmt.Errorf("couldn't calculate embedding, size: %d", len(str))
		}
		log.Debugf("embedding size for %s: %d", sec.Title, len(str))
		info.Sections[i].EmbeddingVec = embResp.Embeddings[0]
	}
	return nil
}

func (info *Information) Render(args ...any) (ret string, err error) {
	funcMap := sprig.FuncMap()
	tmpl := templates.RenderInfo
	for _, arg := range args {
		switch t := arg.(type) {
		case string:
			tmpl = t
		case map[string]func(string) string:
			for key, val := range t {
				funcMap[key] = val
			}
		}
	}
	template, err := template.New("RenderInformation").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		log.Warnf("error %s for template %s: \n", err, tmpl)
	}
	var buf bytes.Buffer
	if err := template.Execute(&buf, info); err != nil {
		return "", err
	}
	return strings.Replace(buf.String(), "\n\n", "\n", -1), nil
}
