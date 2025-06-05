package database

import (
	"bytes"
	"os"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/charmbracelet/log"
	"github.com/openSUSE/kowalski/internal/pkg/file"
	"github.com/openSUSE/kowalski/internal/pkg/templates"
	"github.com/spf13/viper"
)

type PromptInfo struct {
	Name    string
	Version string
	Task    string
	Context string
}

func (kn Knowledge) GetContext(msg string, collections []string, location file.Location, maxSize int) (ret string, err error) {
	if len(collections) == 0 {
		collections, err = kn.GetCollections()
		if err != nil {
			return "", err
		}
	}
	log.Debugf("creating context(%d) for '%s' in '%s'\n", maxSize, msg, collections)
	promptInfo := GetSystemInfo()
	promptInfo.Task = msg
	funcMap := sprig.FuncMap()
	var buf bytes.Buffer
	sysinfo, err := template.New("sysinfo").Funcs(funcMap).Parse(templates.Prompt)
	if err != nil {
		return "", err
	}
	if err = sysinfo.Execute(&buf, promptInfo); err != nil {
		return "", err
	}
	// \TODO just get 5 documents, we can do this dynamically
	infos, err := kn.GetInfos(msg, collections, 5)
	if err != nil {
		return "", err
	}
	contextSize := buf.Len()
	renderedCont := ""
	for _, info := range infos {
		renderedCont += "This help document may be related to the problem:\n"
		if str, err := info.Section.Render(map[string]func(string) string{
			"FileInfo": func(input string) string { return location.Get(input) },
		}); err == nil {
			renderedCont += str
		} else {
			return renderedCont, err
		}
		// check for renderedCont window
		if len(renderedCont)+4*contextSize > maxSize {
			log.Infof("stopped generating context %d > %d", len(renderedCont)+4*contextSize, maxSize)
			break
		}
	}
	buf.Reset()
	sysinfo, err = sysinfo.Parse(templates.Prompt)
	if err != nil {
		return "", err
	}
	promptInfo.Context = renderedCont
	if err = sysinfo.Execute(&buf, promptInfo); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func GetSystemInfo() (sysinfo PromptInfo) {
	osRel := viper.New()
	osRel.SetConfigType("env")
	osRel.SetDefault("NAME", "Unknown linux")
	osRel.SetDefault("VERSION", "0")
	if fh, err := os.Open("/etc/os-release"); err == nil {
		osRel.ReadConfig(fh)
	}
	return PromptInfo{
		Name:    osRel.GetString("NAME"),
		Version: osRel.GetString("VERSION"),
	}
}
