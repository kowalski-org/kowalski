package database

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"text/template"
	"unicode/utf8"

	"github.com/Masterminds/sprig/v3"
	"github.com/openSUSE/kowalski/internal/pkg/templates"
	"github.com/spf13/viper"
)

type SystemInfo struct {
	Name    string
	Version string
}

const maxdirentries = 10
const filemaxsize = 2048

func (kn Knowledge) GetContext(msg string, collections []string, contextSize int) (context string, err error) {
	funcMap := template.FuncMap{}
	for key, value := range sprig.TxtFuncMap() {
		funcMap[key] = value
	}
	var buf bytes.Buffer
	sysinfo, err := template.New("sysinfo").Funcs(funcMap).Parse(templates.SystemPrompt)
	if err := sysinfo.Execute(&buf, GetSystemInfo()); err != nil {
		return "", err
	}
	context += buf.String()
	infos, err := kn.GetInfos(msg, collections)
	if err != nil {
		return "", err
	}
	for _, info := range infos {
		context += "This help document may be related to the problem:\n"
		context += info.Section.Render()
		if len(info.Files) != 0 {
			context += "On the actual system we have following files and directories:"
			for _, file := range info.Files {
				fileStat, err := os.Stat(file)
				if errors.Is(err, os.ErrNotExist) {
					if strings.HasSuffix(file, "/") {
						context += fmt.Sprintf("* directory %s doesn't exist on the system\n", file)
					} else {
						context += fmt.Sprintf("* file %s doesn't exist on the system\n", file)
					}
				}
				if fileStat.IsDir() {
					entries, _ := os.ReadDir(file)
					if len(entries) < maxdirentries {
						var strEnt []string
						for _, ent := range entries {
							strEnt = append(strEnt, ent.Name())
						}
						context += fmt.Sprintf("* directory %s has following entries %s", file, strings.Join(strEnt, ","))
					} else {
						context += fmt.Sprintf("* directory %s has more than %d entries", file, maxdirentries)

					}
				} else {
					if fileStat.Size() < filemaxsize {
						readFile, err := os.Open(os.Args[1])
						if err != nil {
							return "", err
						}
						fileScanner := bufio.NewScanner(readFile)
						fileScanner.Split(bufio.ScanLines)
						fileScanner.Scan()
						if utf8.ValidString(fileScanner.Text()) {
							context += fmt.Sprintf("* file %s has following content: ```\n%s", file, fileScanner.Text())
							for fileScanner.Scan() {
								context += fileScanner.Text()
							}
						}
					} else {
						context += fmt.Sprintf("* file %s exists and larger than %d bytes", file, filemaxsize)
					}
				}
			}
		}
		// check for context window
		if len(context)+4*len(msg) > contextSize {
			break
		}
	}
	return
}

func GetSystemInfo() (sysinfo SystemInfo) {
	osRel := viper.New()
	osRel.SetConfigType("env")
	osRel.SetDefault("NAME", "Unknown linux")
	osRel.SetDefault("VERSION", "0")
	if fh, err := os.Open("/etc/os-release"); err == nil {
		osRel.ReadConfig(fh)
	}
	return SystemInfo{
		Name:    osRel.GetString("NAME"),
		Version: osRel.GetString("VERSION"),
	}
}
