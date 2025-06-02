package information

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"

	"github.com/charmbracelet/log"
	"gopkg.in/yaml.v3"
)

/*
Highly specialzed information is needed for SUSE specfic tools, like
zypper. This information can be provdided in this xml form and has
additionaly to normal documentation the possibilty to add alternative
titles so that a vector db has wide range of arguments
*/

type Curated struct {
	Id       string   `yaml:"Id"`                 // Id for this currated information
	Aliases  []string `yaml:"Aliases,omitempty"`  // alternative titles
	Text     string   `yaml:"Text"`               // one text field for desribing the command
	Commands []string `yaml:"Commands,omitempty"` // commands in the curated text
	Files    []string `yaml:"Files,omitempty"`    // files referenced in this infor
}

func ReadCurated(fileName string) (info Information, err error) {
	info.Source = fileName
	filecont, err := os.ReadFile(fileName)
	if err != nil {
		return info, err
	}
	curratedInfo := Curated{}
	err = yaml.Unmarshal(filecont, &curratedInfo)
	if err != nil {
		return
	}
	hasher := sha256.New()
	io.Copy(hasher, bytes.NewReader(filecont))
	info.Hash = hex.EncodeToString(hasher.Sum(nil))
	log.Debugf("file: %s hash: %s", fileName, info.Hash)
	info.Sections = append(info.Sections, Section{
		Title: curratedInfo.Id,
		Lines: []Line{{
			Text: curratedInfo.Text,
			Type: Text,
		}},
		Commands: curratedInfo.Commands,
		Files:    curratedInfo.Files,
	})
	for _, alt := range curratedInfo.Aliases {
		info.Sections = append(info.Sections, Section{
			Title:   alt,
			IsAlias: true,
		})
	}
	info.Commands = curratedInfo.Commands
	info.Files = curratedInfo.Files
	return info, err
}
