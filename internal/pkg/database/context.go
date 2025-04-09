package database

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

const maxentries = 10
const filemaxsize = 2048

func GetContext(msg string, collections []string) (context string, err error) {
	db, err := New()
	defer db.Close()
	if err != nil {
		return "", err
	}
	infos, err := db.GetInfos(msg, collections)
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
					if len(entries) < maxentries {
						var strEnt []string
						for _, ent := range entries {
							strEnt = append(strEnt, ent.Name())
						}
						context += fmt.Sprintf("* directory %s has following entries %s", file, strings.Join(strEnt, ","))
					} else {
						context += fmt.Sprintf("* directory %s has more than %d entries", file, maxentries)

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
	}
	return
}
