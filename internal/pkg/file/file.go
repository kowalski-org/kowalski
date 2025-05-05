// use a same API for accessing remote and local files
package file

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/log"
)

const (
	maxdirentries = 10
	filemaxsize   = 512
)

type Location interface {
	Get(string) string
}

type Local struct {
	Chroot string
}

type SSH struct {
	Host string
}

// use for mocking, content is just a map with the content of path
type Mock struct {
	Content map[string]string
}

/*
func (loc *Location) Get(path string) string {
	localLoc := Local{
		Chroot: "",
	}
	return localLoc.Get(path)
}
*/

func (loc Local) Get(path string) (out string) {
	path = filepath.Join(loc.Chroot, path)
	fileStat, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		if strings.HasSuffix(path, "/") {
			out += fmt.Sprintf("* directory %s doesn't exist on the system\n", path)
		} else {
			out += fmt.Sprintf("* path %s doesn't exist on the system\n", path)
		}
	}
	if fileStat.IsDir() {
		entries, _ := os.ReadDir(path)
		if len(entries) < maxdirentries {
			var strEnt []string
			for _, ent := range entries {
				strEnt = append(strEnt, ent.Name())
			}
			out += fmt.Sprintf("* directory %s has following entries %s", path, strings.Join(strEnt, ","))
		} else {
			out += fmt.Sprintf("* directory %s has more than %d entries", path, maxdirentries)

		}
	} else {
		if fileStat.Size() < filemaxsize {
			readFile, _ := os.Open(os.Args[1])
			if err != nil {
				out += fmt.Sprintf("* path %s couldn't be opened", path)
			}
			fileScanner := bufio.NewScanner(readFile)
			fileScanner.Split(bufio.ScanLines)
			fileScanner.Scan()
			if utf8.ValidString(fileScanner.Text()) {
				out += fmt.Sprintf("* path %s has following content: ```\n%s", path, fileScanner.Text())
				for fileScanner.Scan() {
					out += fileScanner.Text()
				}
			}
		} else {
			out += fmt.Sprintf("* path %s exists and larger than %d bytes", path, filemaxsize)
		}
	}
	return out
}

func (mock Mock) Get(path string) string {
	if val, ok := mock.Content[path]; ok {
		return val
	} else {
		log.Warnf("couldn't find mock content for path: %s", path)
		return ""
	}
}
