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
				final = new + ".mp4"
				break
			}
		}
	}
	return
}

// Determines ouput filename with .mp4 attached, if file exists, it returns new name
func ParseName(path string) (filename string) {
	filename = path
	if i := strings.LastIndex(path, string(os.PathListSeparator)); i != -1 {
		filename = filename[i+1:]
	}
	if i := strings.LastIndex(path, "."); i != -1 {
		filename = filename[:i]
	}
	filename = fileCheck(filename)
	return
}
