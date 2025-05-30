package docbook

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/beevik/etree"
	"github.com/charmbracelet/log"
	"github.com/openSUSE/kowalski/internal/pkg/information"
)

var entities = map[string]string{
	"nbsp":        " ",
	"prompt.sudo": "sudo ",
	"prompt.user": "",
}

func ParseDocBook(filename string) (info information.Information, err error) {
	doc := etree.NewDocument()
	doc.ReadSettings = etree.ReadSettings{
		Entity: entities,
	}
	// read file handle
	fileHandle, err := os.Open(filename)
	if err != nil {
		return info, err
	}
	defer fileHandle.Close()
	info.Source = filename
	hasher := sha256.New()
	if _, err := io.Copy(hasher, fileHandle); err != nil {
		return info, err
	}
	info.Hash = hex.EncodeToString(hasher.Sum(nil))
	for {
		// read file again as error handling seems better so
		err = doc.ReadFromFile(filename)
		if err != nil {
			errorRegEx := regexp.MustCompile(`XML syntax error on line [0-9]+: invalid character entity &(.*);`)
			match := errorRegEx.FindStringSubmatch(err.Error())
			if len(match) == 2 {
				if match[1] != "" {
					entities[match[1]] = match[1]
				}
			} else {
				log.Warn("other error")
				return info, err
			}
		} else {
			doc = etree.NewDocument()
			doc.ReadSettings = etree.ReadSettings{
				Entity: entities,
			}
			err = doc.ReadFromFile(filename)
			if err != nil {
				return info, fmt.Errorf("couldn't read document %s: %s", filename, err)
			}
			break
		}
	}
	lines := parse(&doc.Element)
	info.Sections = append(info.Sections, information.Section{
		Title: filename,
	})
	for _, line := range lines {
		switch line.Type {
		default:
			info.Sections[len(info.Sections)-1].Lines =
				append(info.Sections[len(info.Sections)-1].Lines, line)
		case information.File:
			// add to explicit file slice, to the lines and slice of info
			info.Sections[len(info.Sections)-1].Lines =
				append(info.Sections[len(info.Sections)-1].Lines, line)
			info.Sections[len(info.Sections)-1].Files =
				append(info.Sections[len(info.Sections)-1].Files, line.Text)
			info.Files = append(info.Files, line.Text)
		case information.Command:
			// same as for file
			info.Sections[len(info.Sections)-1].Lines =
				append(info.Sections[len(info.Sections)-1].Lines, line)
			info.Sections[len(info.Sections)-1].Commands =
				append(info.Sections[len(info.Sections)-1].Commands, line.Text)
			info.Commands = append(info.Commands, line.Text)
		case information.Title:
			info.Sections = append(info.Sections, information.Section{
				Title: line.Text,
			})
		}
	}
	return
}

func parse(elem *etree.Element) (lines []information.Line) {
	for _, e := range elem.ChildElements() {
		switch strings.ToLower(e.Tag) {
		default:
			lines = appendText(lines, e.Text(), e.Tag)
			lines = append(lines, parse(e)...)
			lines = appendText(lines, e.Tail(), e.Parent().Tag)
		case "command", "screen":
			cmdLine := information.Line{
				Type: information.Command,
			}
			buf := []string{cleanStr(e.Text())}
			for _, subCmd := range parse(e) {
				buf = append(buf, subCmd.Text)
			}
			cmdLine.Text = strings.Join(buf, " ")
			lines = append(lines, cmdLine)
			lines = appendText(lines, e.Tail(), "text")
		case "title":
			titleLine := information.Line{
				Type: information.Title,
			}
			switch strings.ToLower(e.Parent().Tag) {
			case "sect2":
				titleLine.Type = information.SubTitle
			case "sect3", "example", "tip", "note":
				titleLine.Type = information.SubSubTitle
			case "warning":
				titleLine.Type = information.Warning
			}
			buf := []string{cleanStr(e.Text())}
			for _, subCmd := range parse(e) {
				buf = append(buf, subCmd.Text)
			}
			titleLine.Text = strings.Join(buf, " ")
			lines = append(lines, titleLine)
			lines = appendText(lines, e.Tail(), "text")
		case "remark", "info":
			continue
		}
	}
	return deformat(lines)
}

// make it global as called several times
var space = regexp.MustCompile(`\s+`)

func cleanStr(input string) string {
	return strings.TrimSpace(space.ReplaceAllString(input, " "))
}

func appendText(lines []information.Line, input string, name string) []information.Line {
	if strings.TrimSpace(input) == "" {
		return lines
	} else {
		// return append(lines, Line{Text: input, Type: GetType(name)})
		return append(lines, information.Line{Text: cleanStr(input), Type: information.GetType(name)})
	}
}

// parse through lines so that e.g. emphasize does not have an own line
func deformat(input []information.Line) (output []information.Line) {
	for i, line := range input {
		switch line.Type {
		default:
			output = append(output, line)
		case "formatted":
			if len(output) > 0 {
				output[len(output)-1].Text += " `" + line.Text + "`"
			} else {
				output = append(output, line)
			}
		case "text":
			if i > 1 && input[i-1].Type == "formatted" && len(output) > 1 {
				output[len(output)-1].Text += " " + line.Text
			} else {
				output = append(output, line)
			}
		}
	}
	return
}
