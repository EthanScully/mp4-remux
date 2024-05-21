package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/EthanScully/mp4-remux/cli/file"
	ffmpeg "github.com/EthanScully/mp4-remux/lib"
)

func excName() (str string) {
	str = os.Args[0]
	if strings.Contains(str, string(os.PathSeparator)) {
		slice := strings.Split(str, string(os.PathSeparator))
		str = slice[len(slice)-1]
	}
	return
}
func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <video-path>\n", excName())
		return
	}
	path := os.Args[1]
	name, err := file.ParseName(path)
	if err != nil {
		panic(err)
	}
	err = ffmpeg.Remux(path, name)
	if err != nil {
		panic(err)
	}
}
