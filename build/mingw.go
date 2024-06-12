package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
)

func getmingw() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	var arch string
	switch runtime.GOARCH {
	case "amd64":
		arch = "x86_64"
	case "arm64":
		arch = "aarch64"
	}
	url := "https://api.github.com/repos/mstorsjo/llvm-mingw/releases/latest"
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != 200 {
		return
	}
	resp.Body.Close()
	var api any
	err = json.Unmarshal(data, &api)
	if err != nil {
		return
	}
	var name string
	assets := api.(map[string]any)["assets"].([]any)
	for _, v := range assets {
		name = v.(map[string]any)["name"].(string)
		if strings.Contains(name, "ucrt") && strings.Contains(name, "ubuntu") && strings.Contains(name, arch) {
			url = v.(map[string]any)["browser_download_url"].(string)
			break
		}
	}
	buildDir := fmt.Sprintf("%s/bin", workingDir)
	err = os.Mkdir(buildDir, 0770)
	if err != nil && !strings.Contains(err.Error(), "exists") {
		return
	}
	defer os.Chdir(workingDir)
	err = os.Chdir(buildDir)
	if err != nil {
		return
	}
	resp, err = http.DefaultClient.Get(url)
	if err != nil {
		return
	}
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	data = make([]byte, 1e7)
	for {
		n, err := resp.Body.Read(data)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		file.Write(data[:n])
	}
	resp.Body.Close()
	file.Close()
	err = sendCmd("tar", "-xf", name)
	if err != nil {
		return
	}
	err = sendCmd("rm", "-rf", name)
	if err != nil {
		return
	}
	name = strings.Replace(name, ".tar.xz", "", 1)
	err = os.Chdir(buildDir + "/" + name)
	if err != nil {
		return
	}
	files, err := os.ReadDir(".")
	if err != nil {
		return
	}
	for _, v := range files {
		sendCmd("cp", "-r", v.Name(), "/usr/local/")
	}
	err = os.Chdir(buildDir)
	if err != nil {
		return
	}
	err = sendCmd("rm", "-rf", name)
	if err != nil {
		return
	}
	return
}
