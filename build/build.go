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
	OS, ARCH, AR, CC, output string
	ffmpegTag                = "n7.0"
	workingDir, _            = os.Getwd()
	makeArgs                 = []string{
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
	--os		windows|linux
	--arch		arm64
	--cc		specify C compiler
	--debug		enable ffmpeg debug mode
	--output	executable output path/name
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
	OS = runtime.GOOS
	ARCH = runtime.GOARCH
	for i, v := range os.Args {
		switch v[:2] {
		case "--":
			switch v[2:] {
			case "os":
				OS = Arg(i + 1)
			case "arch":
				ARCH = Arg(i + 1)
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
	// Make Setup
	switch OS {
	case "linux":
		makeArgs = append([]string{"target-os=linux"}, makeArgs...)
		runtimeArch := runtime.GOARCH
		switch ARCH {
		case runtimeArch:
			CC = "gcc"
			AR = "ar"
		case "amd64":
			CC = "x86_64-linux-gnu-gcc"
			AR = "x86_64-linux-gnu-ar"
			makeArgs = append([]string{"cross-prefix=x86_64-linux-gnu-"}, makeArgs...)
		case "arm64":
			AR = "aarch64-linux-gnu-ar"
			CC = "aarch64-linux-gnu-gcc"
			makeArgs = append([]string{"cross-prefix=aarch64-linux-gnu-"}, makeArgs...)
		default:
			return fmt.Errorf("amd64 and arm64 support only")
		}
	case "windows":
		makeArgs = append([]string{"target-os=mingw32"}, makeArgs...)
		switch ARCH {
		case "amd64":
			AR = "x86_64-w64-mingw32-ar"
			CC = "x86_64-w64-mingw32-gcc"
			makeArgs = append([]string{"cross-prefix=x86_64-w64-mingw32-"}, makeArgs...)
		case "arm64":
			AR = "aarch64-w64-mingw32-ar"
			CC = "aarch64-w64-mingw32-gcc"
			makeArgs = append([]string{"cross-prefix=aarch64-w64-mingw32-"}, makeArgs...)
		default:
			return fmt.Errorf("amd64 and arm64 support only")
		}
	default:
		return fmt.Errorf("linux and windows support only")
	}
	switch ARCH {
	case "arm64":
		makeArgs = append([]string{"arch=aarch64"}, makeArgs...)
	case "amd64":
		makeArgs = append([]string{"arch=x86_64"}, makeArgs...)
	}
	if OS != runtime.GOOS || ARCH != runtime.GOARCH {
		makeArgs = append([]string{"enable-cross-compile"}, makeArgs...)
	}
	// Set Environment Variables
	os.Setenv("GOOS", OS)
	os.Setenv("GOARCH", ARCH)
	os.Setenv("CC", CC)
	os.Setenv("AR", AR)
	os.Setenv("CGO_ENABLED", "1")
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
	sourceDir := fmt.Sprintf("%s/source/", workingDir)
	// Create Build Directory
	buildDir := fmt.Sprintf("%s/build/FFmpeg/%s-%s/", sourceDir, OS, ARCH)
	err = os.MkdirAll(buildDir, 0770)
	if err != nil {
		return err
	}
	// Create Libriary Directory
	libDir := fmt.Sprintf("%s/lib/%s-%s/", workingDir, OS, ARCH)
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

	err = sendCmd("make", fmt.Sprintf("-j%d", runtime.NumCPU()))
	if err != nil {
		return fmt.Errorf("make failed: %s", err)
	}
	err = os.MkdirAll(libDir, 0770)
	if err != nil {
		return err
	}
	err = sendCmd("make", "install")
	if err != nil {
		sendCmd("rm", "-rf", libDir)
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
	_, err = os.Stat(fmt.Sprintf("%s/lib/%s-%s/", workingDir, OS, ARCH))
	if err != nil {
		err := buildFFmpeg()
		if err != nil {
			panic(err)
		}
	}
	ouputPath := fmt.Sprintf("mp4-remux-%s-%s", OS, ARCH)
	if OS == "windows" {
		ouputPath = ouputPath + ".exe"
	}
	if output != "" {
		ouputPath = output
	}
	err = sendCmd("go", "build", "-ldflags=-s -w", "-o", ouputPath, workingDir+"/cli/main.go")
	if err != nil {
		fmt.Println("look at lib/ffmpeg.go for any lib requirments")
		panic(err)
	}
}
