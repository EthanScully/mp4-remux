working=$(pwd)
mkdir -p $working/source/build/ffmpeg/linux/
cd $working/source/build/ffmpeg/linux/
mkdir $working/lib/linux/
CC=gcc
$working/source/FFmpeg/configure \
	--prefix=$working/lib/linux/ \
#	--disable-debug 
make install