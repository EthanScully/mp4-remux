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
go run build/build.go
go run build/build.go --arch arm64
go run build/build.go --os windows
go run build/build.go --os windows --arch arm64
```
### Docker
```BASH
docker run --rm \
        -v $(pwd):/mnt/ -w /mnt/ \
        -t debian:sid bash -c '
        apt update
        apt upgrade -y
        apt install build-essential make nasm yasm zlib1g-dev liblzma-dev golang mingw-w64 gcc-aarch64-linux-gnu tar wget -y
        go run build/build.go
        go run build/build.go --arch arm64
        go run build/build.go --os windows
        wget $(go run build/build.go llvm-mingw)
        tar -xf $(go run build/build.go llvm-mingw name)
        mv $(go run build/build.go llvm-mingw file)/* /usr/local/
        go run build/build.go --os windows --arch arm64'
```