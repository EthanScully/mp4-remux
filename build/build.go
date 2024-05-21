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
	OS, ARCH, CC, output string
	ffmpegTag            = "n7.0"
	workingDir, _        = os.Getwd()
	makeArgs             = []string{
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
	--os	windows|linux
	--arch	arm64
	--cc	specify C compiler
	--debug	enable ffmpeg debug mode
	--output executable output path/name
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
func preCheck() (err error) {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("unsupported build os, linux only")
	}
	_, err = os.Stat("build/build.go")
	if err != nil {
		return fmt.Errorf("this must be run from the root of the repository")
	}
	// parse arguments
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
				ret := -1
				for i, v := range makeArgs {
					if v == "disable-debug" {
						ret = i
					}
				}
				if ret != -1 {
					makeArgs = append(makeArgs[:ret], makeArgs[ret+1:]...)
				}
			case "help":
				fmt.Println(help())
			case "output":
				output = Arg(i + 1)
			}
		}
	}
	// OS check
	switch OS {
	case "linux", "windows":
		break
	default:
		return fmt.Errorf("suported os:\n\twindows\n\tlinux")
	}
	// ARCH check
	switch ARCH {
	case "amd64":
		break
	default:
		return fmt.Errorf("supported arch: amd64")
	}
	// Compiler check
	if CC == "" {
		CC = os.Getenv("CC")
	}
	if CC != os.Getenv("CC") {
		os.Setenv("CC", CC)
	}
	// Make Setup
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
	return
}
func buildFFmpeg() (err error) {
	// Soft Dependency Check
	ok := commandExists(CC)
	if !ok {
		return fmt.Errorf("%s not found", CC)
	}
	ok = commandExists("make")
	if !ok {
		return fmt.Errorf("make not found")
	}
	ok = commandExists("git")
	if !ok {
		return fmt.Errorf("git not found")
	}
	// Git Clone
	defer os.Chdir(workingDir)
	sourceDir := fmt.Sprintf("%s/source/", workingDir)
	err = os.MkdirAll(sourceDir, 0770)
	if err != nil {
		return err
	}
	_, err = os.Stat(sourceDir + "/FFmpeg/")
	if err != nil {
		os.Chdir(sourceDir)
		err = sendCmd("git", "clone", "https://github.com/FFmpeg/FFmpeg/")
		if err != nil {
			return err
		}
	}
	// Git Tag FFmpeg tag: n7.0
	os.Chdir(sourceDir + "/FFmpeg/")
	err = sendCmd("git", "checkout", ffmpegTag)
	if err != nil {
		return err
	}
	// Create Build Directory
	buildDir := fmt.Sprintf("%s/build/ffmpeg/%s/", sourceDir, OS)
	err = os.MkdirAll(buildDir, 0770)
	if err != nil {
		return err
	}
	// Create Libriary Directory
	libDir := fmt.Sprintf("%s/lib/%s/", workingDir, OS)
	// Set output of Make
	makeArgs = append([]string{"prefix=" + libDir}, makeArgs...)
	// Build FFmpeg
	os.Chdir(buildDir)
	var cmd []string
	for _, v := range makeArgs {
		cmd = append(cmd, "--"+v)
	}
	err = sendCmd(sourceDir+"/FFmpeg/configure", cmd...)
	if err != nil {
		return fmt.Errorf("make configure failed: %s\nnasm/yasm might need to be installed", err)
	}
	err = os.MkdirAll(libDir, 0770)
	if err != nil {
		return err
	}
	err = sendCmd("make", "install")
	if err != nil {
		return fmt.Errorf("make install failed: %s", err)
	}
	return
}
func main() {
	err := preCheck()
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = os.Stat(fmt.Sprintf("%s/lib/%s/", workingDir, OS))
	if err != nil {
		err := buildFFmpeg()
		if err != nil {
			panic(err)
		}
	}
	ouputPath := "mp4-remux"
	if output != "" {
		ouputPath = output
	}
	switch OS {
	case "linux":
		err = sendCmd("go", "build", "-ldflags=-s -w", "-o", ouputPath, workingDir+"/cli/main.go")
	case "windows":
		os.Setenv("CGO_ENABLED", "1")
		os.Setenv("GOOS", "windows")
		err = sendCmd("go", "build", "-ldflags=-s -w", "-o", ouputPath, workingDir+"/cli/main.go")
	}
	if err != nil {
		fmt.Println("look at lib/ffmpeg.go for any lib requirments")
		panic(err)
	}
}
