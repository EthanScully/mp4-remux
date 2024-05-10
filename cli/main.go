package main

// build:
// windows:		CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows go build -ldflags="-s -w" -o /mnt/c/Users/scull/Desktop/remux.exe remuxer/cli
// linux:		go build -ldflags="-s -w" -o /mnt/c/Users/scull/Desktop/remux remuxer/cli

import (
	"fmt"
	"os"
	ffmpeg "remuxer/lib"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ffmpeg <video-path>")
		return
	}
	path := os.Args[1]
	name, err := ffmpeg.ParseName(path)
	if err != nil {
		panic(err)
	}
	err = ffmpeg.Remux(path, name)
	if err != nil {
		panic(err)
	}
}
