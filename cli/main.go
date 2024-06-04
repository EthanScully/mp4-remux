package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/EthanScully/mp4-remux/cli/file"
	ffmpeg "github.com/EthanScully/mp4-remux/lib"
)

func main() {
	if len(os.Args) < 2 {
		str := os.Args[0]
		if i := strings.LastIndex(str, string(os.PathSeparator)); i != -1 {
			str = str[i+1:]
		}
		fmt.Printf("Usage: %s <video-path>\n", str)
		return
	}
	path := os.Args[1]
	filename := file.ParseName(path)
	err := ffmpeg.Remux(path, filename)
	if err != nil {
		panic(err)
	}
}
