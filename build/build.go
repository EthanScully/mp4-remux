package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var (
	workingDir, _ = os.Getwd()
	makeArgs      = []string{
		"disable-avdevice",
		"disable-postproc",
		"disable-avfilter",
		"disable-doc",
		"disable-htmlpages",
		"disable-manpages",
		"disable-podpages",
		"disable-txtpages",
		"disable-programs",
		"disable-network",
		"disable-everything",
		"enable-demuxer=mov",
		"enable-demuxer=matroska",
		"enable-demuxer=m4v",
		"enable-demuxer=mpegts",
		"enable-muxer=mp4",
		"enable-decoder=aac",
		"enable-parser=aac",
		"enable-parser=ac3",
		"enable-parser=hdr",
		"enable-parser=av1",
		"enable-parser=hevc",
		"enable-parser=h264",
		"enable-protocol=file",
		"disable-debug",
	}
)

func Arg(i int) (arg string) {
	if len(os.Args) > i {
		arg = os.Args[i]
	}
	return
}
func help() (out string) {
	out = `Supported Arguments:
	--os
	--arch
	--cc
	--debug
	--help
`
	return
}
func sendCmd(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	err := cmd.Start()
	if err != nil {
		return err
	}
	streamOutput := func(reader io.Reader) {
		buffer := make([]byte, 1024)
		for {
			n, err := reader.Read(buffer)
			if err != nil {
				if err != io.EOF {
					fmt.Println("Error reading output:", err)
				}
				break
			}
			if n > 0 {
				fmt.Print(string(buffer[:n]))
			}
		}
	}
	go streamOutput(stdout)
	go streamOutput(stderr)
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}
func commandExists(cmd string) bool {
	output, err := exec.Command("which", cmd).CombinedOutput()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) != ""
}
func searchArray(array []string, str string) (index int) {
	index = -1
	for i, v := range array {
		if v == str {
			index = i
		}
	}
	return
}
func main() {
	if runtime.GOOS != "linux" {
		fmt.Println("Unsupported Build OS, Linux Only")
		return
	}
	_, err := os.Stat("build/build.go")
	if err != nil {
		fmt.Println("this must be run from the root of the repository")
		return
	}
	var OS, ARCH, CC string
	for i, v := range os.Args {
		switch v[:2] {
		case "--":
			switch v[2:] {
			case "os":
				OS = Arg(i + 1)
			case "arch":
				ARCH = Arg(i + 1)
			case "cc":
				CC = Arg(i + 1)
			case "debug":
				ret := searchArray(makeArgs, "disable-debug")
				if ret != -1 {
					makeArgs = append(makeArgs[:ret], makeArgs[ret+1:]...)
				}
			case "help":
				fmt.Println(help())
			}
		}
	}
	switch OS {
	case "linux", "windows":
		break
	default:
		fmt.Println("Suported OS:\n\twindows\n\tlinux")
		return
	}
	switch ARCH {
	case "amd64":
		break
	default:
		fmt.Println("Unsupported ARCH")
		fmt.Println("Supported ARCH:\n\tamd64")
		return
	}
	if CC == "" {
		CC = os.Getenv("CC")
	}
	if CC != os.Getenv("CC") {
		os.Setenv("CC", CC)
	}
	switch OS {
	case "linux":
		if CC == "" {
			CC = "gcc"
		}
	case "windows":
		if CC == "" || CC == "gcc" {
			fmt.Println("CC environment variable for windows compilation not specified")
			return
		}
		makeArgs = append([]string{"enable-cross-compile"}, makeArgs...)
		makeArgs = append([]string{"target-os=mingw32"}, makeArgs...)
		switch ARCH {
		case "amd64":
			makeArgs = append([]string{"cross-prefix=x86_64-w64-mingw32-"}, makeArgs...)
			makeArgs = append([]string{"arch=x86_64"}, makeArgs...)
		}
	}
	defer os.Chdir(workingDir)
	libDir := fmt.Sprintf("%s/lib/%s/", workingDir, OS)
	err = os.MkdirAll(libDir, 0770)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	makeArgs = append([]string{"prefix="+libDir}, makeArgs...)
	ok := commandExists(CC)
	if !ok {
		fmt.Println("CC not found")
		return
	}
	ok = commandExists("make")
	if !ok {
		fmt.Println("Make not found")
		return
	}
	ok = commandExists("git")
	if !ok {
		fmt.Println("Git not found")
		return
	}
	os.Chdir(workingDir + "/source/")
	err = sendCmd("git", "clone", "https://github.com/FFmpeg/FFmpeg/")
	if err != nil {
		if err.Error() != "exit status 128" {
			fmt.Println(err)
			return
		}
	}
	os.Chdir(workingDir + "/source/FFmpeg/")
	err = sendCmd("git", "checkout", "n7.0")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	buildDir := fmt.Sprintf("%s/source/build/ffmpeg/%s/", workingDir, OS)
	err = os.MkdirAll(buildDir, 0770)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	os.Chdir(buildDir)
	var cmd []string
	for _, v := range makeArgs {
		cmd = append(cmd, "--"+v)
	}
	err = sendCmd(workingDir+"/source/FFmpeg/configure", cmd...)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = sendCmd("make", "install")
	if err != nil {
		fmt.Println(err)
		return
	}

}
