package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

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
	if os.Args[1][len(os.Args[1])-1] == '*' {
		batch(os.Args[1][:len(os.Args[1])-1])
		return
	} else {
		run(os.Args[1])
		return
	}
}
func batch(directory string) {
	prefix := directory
	directory = directory[:len(directory)-1]
	dir, err := os.ReadDir(directory)
	if err != nil {
		fmt.Printf("error reading the directory: %s\n", directory)
	}
	var todo []string
	for _, file := range dir {
		if !file.IsDir() {
			filepath := prefix + file.Name()
			todo = append(todo, filepath)
		}
	}
	var wg sync.WaitGroup
	var mtx sync.Mutex
	var running int
	threads := runtime.NumCPU()
	for _, path := range todo {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			mtx.Lock()
			running++
			mtx.Unlock()
			run(path)
			mtx.Lock()
			running--
			mtx.Unlock()
		}(path)
		for running >= threads {
			time.Sleep(time.Microsecond * 250)
		}
	}
	wg.Wait()
}
func run(path string) {
	filename := file.ParseName(path)
	err := ffmpeg.Remux(path, filename)
	if err != nil {
		fmt.Println(err)
	}
}
