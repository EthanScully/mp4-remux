# Mux MP4s with FFmpeg in Go
Supports coverting  mkv | ts | mp4  into an mp4 with the MOOV atom at the start of the file, also known as fast start or web optimized

## Usage
```POWERSHELL
.\mp4-remux-win.exe <filepath>
```
This program utilizes the [FFmpeg](https://github.com/FFmpeg/FFmpeg) C api with CGO to remux the video files 
## Build
only supports building on linux right now

when in repository directory:
```BASH
go run build/build.go --os linux --arch amd64 --output mp4-remux-linux
go run build/build.go --os windows --arch amd64 --output mp4-remux-win.exe
```
### Docker
```BASH
docker run --rm \
        -v $(pwd):/mnt/ -w /mnt/ \
        -t debian:sid bash -c '
        apt update
        apt upgrade -y
        apt install build-essential git make nasm yasm zlib1g-dev liblzma-dev golang mingw-w64 gcc-aarch64-linux-gnu -y
        git config --global --add safe.directory "*"
        go run build/build.go --os linux --arch amd64 --output mp4-remux-linux
        go run build/build.go --os windows --arch amd64 --output mp4-remux-win.exe'
```