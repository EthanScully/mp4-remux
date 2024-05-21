package file

import (
	"fmt"
	"os"
	"strings"
)

func fileCheck(name string, or ...int) (final string, err error) {
	var num int
	if len(or) == 0 {
		num = 0
	} else {
		num = or[0] + 1
	}
	files, err := os.ReadDir(".")
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if num == 0 {
			if file.Name() == fmt.Sprintf("%s.mp4", name) {
				final, err = fileCheck(name, num)
				return
			}
		} else {
			if file.Name() == fmt.Sprintf("%s(%d).mp4", name, num) {
				final, err = fileCheck(name, num)
				return
			}
		}
	}
	if num == 0 {
		final = name + ".mp4"
		return
	} else {
		final = fmt.Sprintf("%s(%d).mp4", name, num)
		return
	}
}
func ParseName(path string) (filename string, err error) {
	if len(path) < 5 {
		err = fmt.Errorf("path length")
		return
	}
	if !strings.Contains(path[len(path)-5:], ".") {
		err = fmt.Errorf("path must contain a file extension")
		return
	}
	extLoc := strings.LastIndex(path, ".")
	if !strings.Contains(path, string(os.PathSeparator)) {
		filename = path[:extLoc]
	} else {
		lastSep := strings.LastIndex(path, string(os.PathSeparator))
		filename = path[lastSep+1 : extLoc]
	}
	filename, err = fileCheck(filename)
	return
}
