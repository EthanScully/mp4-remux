package ffmpeg

/*
#cgo windows,amd64 CFLAGS: -Iwindows-amd64/include
#cgo windows,arm64 CFLAGS: -Iwindows-arm64/include
#cgo linux,amd64 CFLAGS: -Ilinux-amd64/include
#cgo linux,arm64 CFLAGS: -Ilinux-arm64/include
#cgo LDFLAGS: -static
#cgo LDFLAGS: -lavformat -lswscale -lavcodec -lavutil -lswresample
#cgo linux,amd64 LDFLAGS: -Llinux-amd64/lib -lz -lm -llzma
#cgo linux,arm64 LDFLAGS: -Llinux-arm64/lib -lm
#cgo windows,amd64 LDFLAGS: -Lwindows-amd64/lib -lbcrypt
#cgo windows,arm64 LDFLAGS: -Lwindows-arm64/lib -lbcrypt
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libavutil/imgutils.h>
#include <libavutil/mem.h>
#include <libavutil/timestamp.h>
#include <libswscale/swscale.h>
#include <libavutil/log.h>
void av_log_wrapper(void* avcl, int level, char* string) {
	av_log(avcl, level, "%s", string);
}
*/
import "C"
import (
	"fmt"
	"os"
	"slices"
	"unsafe"
)

type packetQueue struct {
	packets []*C.AVPacket
	num     int
	pos     int
	pts     [][]C.int64_t
	done    bool
	last    int
}

func (q *packetQueue) init(num, streams int) (err error) {
	q.num = num
	q.packets = make([]*C.AVPacket, q.num)
	for i := range q.packets {
		q.packets[i] = C.av_packet_alloc()
		if q.packets[i] == nil {
			err = fmt.Errorf("could not allocate packet")
			return
		}
	}
	q.pts = make([][]C.int64_t, streams)
	return
}
func (q *packetQueue) free() {
	for i := range q.packets {
		C.av_packet_free(&q.packets[i])
	}
}
func (q *packetQueue) next() {
	q.pos++
	if q.pos >= q.num {
		q.pos = 0
	}
}

func Remux(filepath, filename string) (err error) {
	// Init
	filepath = "file:" + filepath
	var ifmt_ctx *C.AVFormatContext = nil
	if C.avformat_open_input(&ifmt_ctx, C.CString(filepath), nil, nil) < 0 {
		err = fmt.Errorf("could not open input file \"%s\"", filepath)
		return
	}
	defer C.avformat_close_input(&ifmt_ctx)
	if C.avformat_find_stream_info(ifmt_ctx, nil) < 0 {
		err = fmt.Errorf("failed to retrieve input stream information")
		return
	}
	var ofmt_ctx *C.AVFormatContext
	C.avformat_alloc_output_context2(&ofmt_ctx, nil, nil, C.CString(filename))
	if ofmt_ctx == nil {
		err = fmt.Errorf("could not create output context")
		return
	}
	defer C.avformat_free_context(ofmt_ctx)
	// //
	// Map Streams
	// //
	stream_mapping_size := ifmt_ctx.nb_streams
	var stream_mapping *C.int = nil
	stream_mapping = (*C.int)(C.av_calloc(C.size_t(stream_mapping_size), (C.size_t)(unsafe.Sizeof(*stream_mapping))))
	if stream_mapping == nil {
		err = fmt.Errorf("could not allocate stream mapping")
		return
	}
	defer C.av_freep(unsafe.Pointer(&stream_mapping))
	ofmt := ofmt_ctx.oformat
	var stream_index C.int
	for i := 0; i < int(stream_mapping_size); i++ {
		in_stream := unsafe.Slice(ifmt_ctx.streams, i+1)[i]
		in_codecpar := in_stream.codecpar
		streamMapping := unsafe.Slice(stream_mapping, i+1)
		switch in_codecpar.codec_type {
		case C.AVMEDIA_TYPE_AUDIO, C.AVMEDIA_TYPE_VIDEO:
		default:
			streamMapping[i] = -1
			continue
		}
		if in_codecpar.codec_type == C.AVMEDIA_TYPE_VIDEO {
			switch in_codecpar.codec_id {
			case C.AV_CODEC_ID_AV1, C.AV_CODEC_ID_H264, C.AV_CODEC_ID_HEVC:
			default:
				C.av_log_wrapper(unsafe.Pointer(ifmt_ctx), C.AV_LOG_ERROR, C.CString(fmt.Sprintf("video stream #%d not supported, skipping... \n", i)))
				streamMapping[i] = -1
				continue
			}
		}
		if in_codecpar.codec_type == C.AVMEDIA_TYPE_AUDIO {
			switch in_codecpar.codec_id {
			case C.AV_CODEC_ID_AAC, C.AV_CODEC_ID_AC3, C.AV_CODEC_ID_DTS:
			default:
				C.av_log_wrapper(unsafe.Pointer(ifmt_ctx), C.AV_LOG_ERROR, C.CString(fmt.Sprintf("audio stream #%d not supported, skipping... \n", i)))
				streamMapping[i] = -1
				continue
			}
		}
		streamMapping[i] = stream_index
		stream_index++
		out_stream := C.avformat_new_stream(ofmt_ctx, nil)
		if out_stream == nil {
			err = fmt.Errorf("failed allocating output stream")
			return
		}
		if C.avcodec_parameters_copy(out_stream.codecpar, in_codecpar) < 0 {
			err = fmt.Errorf("failed to copy codec parameters")
			return
		}
		out_stream.codecpar.codec_tag = 0
	}
	// Create Ouput File
	if (ofmt.flags & C.AVFMT_NOFILE) == 0 {
		if C.avio_open(&ofmt_ctx.pb, C.CString(filename), C.AVIO_FLAG_WRITE) < 0 {
			err = fmt.Errorf("could not open output file \"%s\"", filename)
			return
		}
	}
	// Specify faststart
	defer C.avio_closep(&ofmt_ctx.pb)
	var myDict *C.AVDictionary
	C.av_dict_set(&myDict, C.CString("movflags"), C.CString("+faststart"), 0)
	if C.avformat_write_header(ofmt_ctx, &myDict) < 0 {
		err = fmt.Errorf("error occurred when opening output file")
		return
	}
	// Init Packet Queue
	var pq packetQueue
	pq.init(100, int(stream_mapping_size))
	defer pq.free()
	for i := range pq.packets {
		if C.av_read_frame(ifmt_ctx, pq.packets[i]) < 0 {
			pq.done = true
			pq.last = i
			break
		}
		index := pq.packets[i].stream_index
		pq.pts[index] = append(pq.pts[index], pq.packets[i].pts)
	}
	// //
	// dts and pts processing
	// //
	var dtsOffset C.int64_t
	var offset C.int64_t = pq.pts[0][0]
	var ptsOffset C.int64_t
	streamMapping := unsafe.Slice(stream_mapping, stream_mapping_size)
	for i, ptsList := range pq.pts {
		if len(ptsList) == 0 {
			continue
		}
		if streamMapping[i] == -1 {
			continue
		}
		sorted := make([]C.int64_t, len(ptsList))
		copy(sorted, ptsList)
		slices.Sort(sorted)
		for j, pts := range sorted {
			diff := ptsList[j] - pts
			if diff < dtsOffset {
				dtsOffset = diff
			}
		}
		if sorted[0] < offset {
			offset = sorted[0]
		}
	}
	if offset == C.AV_NOPTS_VALUE {
		return fmt.Errorf("first pts is AV_NOPTS_VALUE: %v", pq.pts)
	}
	for i, ptsList := range pq.pts {
		if streamMapping[i] == -1 {
			continue
		}
		slices.Sort(ptsList)
	}
	var sumDiff float64
	var sumDiffCount float64
	for i, v := range pq.pts {
		if streamMapping[i] == -1 {
			continue
		}
		for i := 1; i < len(v); i++ {
			sumDiff += float64(v[i] - v[i-1])
			sumDiffCount++
		}
	}
	avgDiff := C.int64_t(sumDiff / sumDiffCount)
	maxdtsOffset := -10 * avgDiff
	if dtsOffset < maxdtsOffset {
		dtsOffset = maxdtsOffset
	}
	dtsOffset -= offset
	ptsOffset -= offset
	lastdts := make([]C.int64_t, len(pq.pts))
	for i := range lastdts {
		lastdts[i] = -1
	}
	// //
	// muxing
	// //
	read := func() {
		if C.av_read_frame(ifmt_ctx, pq.packets[pq.pos]) < 0 {
			if !pq.done {
				pq.done = true
				pq.last = pq.pos
			}
		} else {
			index := pq.packets[pq.pos].stream_index
			pq.pts[index] = append(pq.pts[index], pq.packets[pq.pos].pts)
			slices.Sort(pq.pts[index])
		}
		pq.next()
	}
	var packet int
	for {
		packet++
		if pq.done && pq.pos == pq.last {
			break
		}
		pkt := pq.packets[pq.pos]
		if streamMapping[pkt.stream_index] == -1 {
			read()
			continue
		}
		pkt.pts += ptsOffset
		pkt.dts = pq.pts[pkt.stream_index][0] + dtsOffset
		pq.pts[pkt.stream_index] = pq.pts[pkt.stream_index][1:]
		if pkt.dts == lastdts[pkt.stream_index] {
			C.av_log_wrapper(unsafe.Pointer(ifmt_ctx), C.AV_LOG_INFO, C.CString("increasing pkt offset by 1\n"))
			pkt.dts++
			pkt.pts++
			ptsOffset++
			dtsOffset++
		}
		if pkt.pts < pkt.dts {
			diff := pkt.dts - pkt.pts
			if diff > avgDiff*10 {
				C.av_log_wrapper(unsafe.Pointer(ifmt_ctx), C.AV_LOG_ERROR, C.CString(fmt.Sprintf("dts > pts, %v:%v, dropping packet\n", pkt.dts, pkt.pts)))
				read()
				continue
			}
			C.av_log_wrapper(unsafe.Pointer(ifmt_ctx), C.AV_LOG_INFO, C.CString(fmt.Sprintf("dts > pts, %v:%v, added pts offset: %v\n", pkt.dts, pkt.pts, diff)))
			ptsOffset += diff
			pkt.pts += diff
		}
		lastdts[pkt.stream_index] = pkt.dts
		err = rescalePacket(pkt, ifmt_ctx, ofmt_ctx, stream_mapping_size, stream_mapping)
		if err != nil {
			read()
			continue
		}
		pkt.pos = -1
		if C.av_interleaved_write_frame(ofmt_ctx, pkt) < 0 {
			err = fmt.Errorf("error muxing packet:%v", pq.pts[pkt.stream_index])
			break
		}
		read()
	}
	C.av_write_trailer(ofmt_ctx)
	C.av_log_wrapper(unsafe.Pointer(ifmt_ctx), C.AV_LOG_VERBOSE, C.CString(fmt.Sprintf("Total Packets: %v, pts & dts offset: %v,%v\n", packet, ptsOffset, dtsOffset)))
	return
}
func rescalePacket(pkt *C.AVPacket, ifmt_ctx, ofmt_ctx *C.AVFormatContext, stream_mapping_size C.uint, stream_mapping *C.int) (err error) {
	var in_stream, out_stream *C.AVStream
	in_stream = unsafe.Slice(ifmt_ctx.streams, pkt.stream_index+1)[pkt.stream_index]
	if pkt.stream_index >= C.int(stream_mapping_size) || unsafe.Slice(stream_mapping, pkt.stream_index+1)[pkt.stream_index] < 0 {
		err = fmt.Errorf("err")
		C.av_packet_unref(pkt)
		return
	}
	pkt.stream_index = unsafe.Slice(stream_mapping, pkt.stream_index+1)[pkt.stream_index]
	out_stream = unsafe.Slice(ofmt_ctx.streams, pkt.stream_index+1)[pkt.stream_index]
	C.av_packet_rescale_ts(pkt, in_stream.time_base, out_stream.time_base)
	return
}
func init() {
	C.av_log_set_level(C.AV_LOG_ERROR)
	if len(os.Args) > 2 {
		switch os.Args[2] {
		case "-v":
			C.av_log_set_level(C.AV_LOG_VERBOSE)
		case "-vv":
			C.av_log_set_level(C.AV_LOG_DEBUG)
		case "-vvv":
			C.av_log_set_level(C.AV_LOG_MAX_OFFSET)
		}
	}
}
