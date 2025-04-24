package docbook

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"html"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/beevik/etree"
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
	fileHandle, err := os.Open("filename")
	if err != nil {
		return info, err
	}
	defer fileHandle.Close()
	info.Source = filename
	hasher := sha256.New()
	if _, err := io.Copy(hasher, fileHandle); err != nil {
		return info, err
	}
	info.Hash = string(hasher.Sum(nil))
	for {
		_, err = doc.ReadFrom(fileHandle)
		if err != nil {
			errorRegEx := regexp.MustCompile(`XML syntax error on line [0-9]+: invalid character entity &(.*);`)
			match := errorRegEx.FindStringSubmatch(err.Error())
			if len(match) == 2 {
				if match[1] != "" {
					entities[match[1]] = match[1]
				}
			} else {
				return info, err
			}
		} else {
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
			info.Sections[len(info.Sections)-1].Lines =
				append(info.Sections[len(info.Sections)-1].Lines, line)
			info.Sections[len(info.Sections)-1].Files =
				append(info.Sections[len(info.Sections)-1].Files, line.Text)
		case information.Command:
			info.Sections[len(info.Sections)-1].Lines =
				append(info.Sections[len(info.Sections)-1].Lines, line)
			info.Sections[len(info.Sections)-1].Commands =
				append(info.Sections[len(info.Sections)-1].Commands, line.Text)
		case information.Title:
			if len(info.Sections) == 1 {
				info.Sections[0].Title = line.Text
			} else {
				info.Sections = append(info.Sections, information.Section{
					Title: line.Text,
				})
			}
		}
	}
	return
}

func cleanText(input string) (output string) {
	output = input
	output = html.UnescapeString(output)
	output = strings.Replace(output, "prompt.sudo", " sudo ", -1)
	output = strings.Replace(output, "nbsp", " ", -1)
	output = strings.Replace(output, "  ", " ", -1)
	output = strings.TrimSpace(output)
	output = strings.Replace(output, "\n\n", "\n", -1)
	return
}

func ReadEntity(filename string) (entities map[string]string, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	entityRegex := regexp.MustCompile(`<!ENTITY\s+([\p{L}][^\s]+)\s+"([^"]+)"\s*>`)

	entities = make(map[string]string)
	for scanner.Scan() {
		line := scanner.Text()
		match := entityRegex.FindStringSubmatch(line)
		if len(match) == 3 {
			if match[1] != "" && match[2] != "" {
				entities[match[1]] = match[2]
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}
	return entities, nil
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
		}
	}
	return deformat(lines)
}

var space = regexp.MustCompile(`\s+`)

func cleanStr(input string) string {
	return strings.TrimSpace(space.ReplaceAllString(input, " "))
}

func appendText(lines []information.Line, input string, name string) []information.Line {
	if strings.TrimSpace(input) == "" {
		return lines
	} else {
		// return append(lines, Line{Text: input, Type: GetType(name)})
		return append(lines, information.Line{Text: cleanStr(input), Type: getType(name)})
	}
}

func getType(str string) information.LineType {
	switch strings.ToLower(str) {
	case "title":
		return information.Title
	case "command", "screen":
		return information.Command
	case "package", "emphasis", "literal", "option", "replaceable":
		return information.Formatted
	default:
		return information.Text
	}
}

// parse through lines so that e.g. emphasize does not have an own line
func deformat(input []information.Line) (output []information.Line) {
	for i, line := range input {
		switch line.Type {
		default:
			output = append(output, line)
		case information.Formatted:
			if len(output) > 0 {
				output[len(output)-1].Text += " `" + line.Text + "`"
			} else {
				output = append(output, line)
			}
		case information.Text:
			if i > 1 && input[i-1].Type == information.Formatted && len(output) > 1 {
				output[len(output)-1].Text += " " + line.Text
			} else {
				output = append(output, line)
			}
		}
	}
	return
}
