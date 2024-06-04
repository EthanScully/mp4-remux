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
go run build/*
go run build/* --arch arm64
go run build/* --os windows
go run build/* --os windows --arch arm64
```
### Docker
```BASH
docker run --rm \
        -v $(pwd):/mnt/ -w /mnt/ \
        -t debian:sid bash -c '
        apt update
        apt install build-essential make nasm yasm zlib1g-dev liblzma-dev golang gcc-aarch64-linux-gnu tar ca-certificates -y
        go run build/* mingw
        go run build/*
        go run build/* --arch arm64
        go run build/* --os windows
        go run build/* --os windows --arch arm64'
```