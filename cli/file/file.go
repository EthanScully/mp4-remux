package file

import (
	"fmt"
	"os"
	"strings"
)

func fileCheck(name string) (final string) {
	final = name + ".mp4"
	if _, err := os.Stat(final); err == nil {
		for i := 1; i > 0; i++ {
			new := fmt.Sprintf("%s(%d)", name, i)
			if _, err := os.Stat(new + ".mp4"); err != nil {
				err = nil
				final = new + ".mp4"
				break
			}
		}
	}
	return
}

// Determines ouput filename with .mp4 attached, if file exists, it returns new name
func ParseName(path string) (filename string, err error) {
	filename = path
	if i := strings.LastIndex(path, string(os.PathListSeparator)); i != -1 {
		filename = path[i+1:]
	}
	if i := strings.LastIndex(path, "."); i != -1 {
		filename = path[:i]
	}
	filename = fileCheck(filename)
	return
}
