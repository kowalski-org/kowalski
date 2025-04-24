package information

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/log"

	"github.com/Masterminds/sprig/v3"
	"github.com/openSUSE/kowalski/internal/app/ollamaconnector"
	"github.com/openSUSE/kowalski/internal/pkg/templates"
)

const maxdirentries = 10
const filemaxsize = 2048

type RenderData struct {
	Level int
	Section
}
)

type Line struct {
	Text string
	Type string
}

type Information struct {
	OS       []string
	Hash     string
	Source   string
	Sections []Section
}

type Section struct {
	Title        string    `yaml:"Title,omitempty"`
	EmbeddingVec []float32 `yaml:"EmbeddingVec,omitempty"`
	Lines        []Line    `yaml:"Lines,omitempty"`
	Files        []string  `yaml:"Files,omitempty"`
	Commands     []string  `yaml:"Commands,omitempty"`
}

type RetSection struct {
	Dist float32
	Section
}


func (info *Section) RenderWithFiles(args ...any) (ret string, err error) {
	fileFunc := map[string]func(string) string{
		"FileContext": func(in string) (out string) {
			if len(info.Files) != 0 {
				out += "On the actual system we have following files and directories:"
				for _, file := range info.Files {
					fileStat, err := os.Stat(file)
					if errors.Is(err, os.ErrNotExist) {
						if strings.HasSuffix(file, "/") {
							out += fmt.Sprintf("* directory %s doesn't exist on the system\n", file)
						} else {
							out += fmt.Sprintf("* file %s doesn't exist on the system\n", file)
						}
					}
					if fileStat.IsDir() {
						entries, _ := os.ReadDir(file)
						if len(entries) < maxdirentries {
							var strEnt []string
							for _, ent := range entries {
								strEnt = append(strEnt, ent.Name())
							}
							out += fmt.Sprintf("* directory %s has following entries %s", file, strings.Join(strEnt, ","))
						} else {
							out += fmt.Sprintf("* directory %s has more than %d entries", file, maxdirentries)

						}
					} else {
						if fileStat.Size() < filemaxsize {
							readFile, _ := os.Open(os.Args[1])
							if err != nil {
								out += fmt.Sprintf("* file %s couldn't be opened", file)
							}
							fileScanner := bufio.NewScanner(readFile)
							fileScanner.Split(bufio.ScanLines)
							fileScanner.Scan()
							if utf8.ValidString(fileScanner.Text()) {
								out += fmt.Sprintf("* file %s has following content: ```\n%s", file, fileScanner.Text())
								for fileScanner.Scan() {
									out += fileScanner.Text()
								}
							}
						} else {
							out += fmt.Sprintf("* file %s exists and larger than %d bytes", file, filemaxsize)
						}
					}
				}
			}
			return
		}}
	return info.Render(fileFunc)
}

func (info *Section) Render(args ...any) (string, error) {
	level := 0
	funcMap := sprig.FuncMap()
	tmpl := templates.RenderInfo
	for _, arg := range args {
		switch t := arg.(type) {
		case string:
			tmpl = t
		case int:
			level = t
		case map[string]func(string) string:
			for key, val := range t {
				funcMap[key] = val
			}
		}
	}
	// funcMap["RenderSubsections"] = info.RenderSubsections
	funcMap["Section"] = func() string {
		return strings.Repeat("#", level)
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
		return "", err
	}
	return strings.Replace(buf.String(), "\n\n", "\n", -1), nil
}

func (info *Information) Empty() bool {
	return len(info.Sections) == 0
}

func (info *Information) CreateEmbedding() (err error) {
	for _, sec := range info.Sections {
		str := sec.Render()
		embResp, err := ollamaconnector.Ollamasettings.GetEmbeddings([]string{str})
		if err != nil {
			return err
		}
		if len(embResp.Embeddings) == 0 {
			return fmt.Errorf("couldn't calculate embedding")
		}
		sec.EmbeddingVec = embResp.Embeddings[0]
	}
	return nil
}

func (info *Information) Render(args ...any) (str string) {
	str = fmt.Sprintf("File: %s\nOS: %v\n", info.Source, info.OS)
	for _, sec := range info.Sections {
		str += sec.Render()
	}
	return

}
