package mux

/*
#cgo pkg-config: libavformat libavcodec libavutil
#include <libavformat/avformat.h>
#include <libavcodec/avcodec.h>
#include <libavutil/avutil.h>
#include <stdlib.h>
#include <string.h>

static char* tt_strdup(const char* s) {
    size_t n = strlen(s) + 1;
    char* out = (char*)malloc(n);
    if (out) memcpy(out, s, n);
    return out;
}

static void tt_set_error(char** err, int code, const char* prefix) {
    char detail[AV_ERROR_MAX_STRING_SIZE];
    av_strerror(code, detail, sizeof(detail));
    char buffer[512];
    snprintf(buffer, sizeof(buffer), "%s: %s", prefix, detail);
    *err = tt_strdup(buffer);
}

static int tt_open_input(AVFormatContext** ctx, const char* path, const char* headers, char** err) {
	AVDictionary* opts = NULL;
	if (headers && headers[0]) av_dict_set(&opts, "headers", headers, 0);
	int ret = avformat_open_input(ctx, path, NULL, &opts);
	av_dict_free(&opts);
    if (ret < 0) {
        tt_set_error(err, ret, "open input");
        return ret;
    }
    ret = avformat_find_stream_info(*ctx, NULL);
    if (ret < 0) {
        tt_set_error(err, ret, "find stream info");
        return ret;
    }
    return 0;
}

static int tt_read_selected_packet(AVFormatContext* input, int stream_index, AVPacket* packet, char** err) {
    int ret;
    while ((ret = av_read_frame(input, packet)) >= 0) {
        if (packet->stream_index == stream_index) return 1;
        av_packet_unref(packet);
    }
    if (ret == AVERROR_EOF) return 0;
    tt_set_error(err, ret, "read packet");
    return ret;
}

static int64_t tt_packet_time(AVPacket* packet, AVRational time_base) {
    int64_t timestamp = packet->pts == AV_NOPTS_VALUE ? packet->dts : packet->pts;
    if (timestamp == AV_NOPTS_VALUE) return INT64_MAX;
    return av_rescale_q(timestamp, time_base, AV_TIME_BASE_Q);
}

static int tt_write_packet(AVFormatContext* output, AVPacket* packet, AVStream* in_stream, AVStream* out_stream, char** err) {
    packet->stream_index = out_stream->index;
    av_packet_rescale_ts(packet, in_stream->time_base, out_stream->time_base);
    int ret = av_interleaved_write_frame(output, packet);
    if (ret < 0) tt_set_error(err, ret, "write packet");
    return ret;
}

static int tt_copy_stream(AVFormatContext* output, AVStream* input_stream, AVStream** output_stream, char** err) {
    AVStream* stream = avformat_new_stream(output, NULL);
    if (!stream) {
        *err = tt_strdup("create output stream: allocation failed");
        return AVERROR(ENOMEM);
    }
    int ret = avcodec_parameters_copy(stream->codecpar, input_stream->codecpar);
    if (ret < 0) {
        tt_set_error(err, ret, "copy codec parameters");
        return ret;
    }
    stream->codecpar->codec_tag = 0;
    stream->time_base = input_stream->time_base;
    *output_stream = stream;
    return 0;
}

static int tt_remux_avformat(const char* video_path, const char* video_headers, const char* audio_path, const char* audio_headers, const char* output_path, char** err) {
    AVFormatContext *video_input = NULL, *audio_input = NULL, *output = NULL;
    AVStream *video_out = NULL, *audio_out = NULL;
    AVPacket *video_packet = NULL, *audio_packet = NULL;
    int ret = 0;
    int video_ready = 0, audio_ready = 0;

	ret = tt_open_input(&video_input, video_path, video_headers, err);
	if (ret < 0) goto cleanup;
	ret = tt_open_input(&audio_input, audio_path, audio_headers, err);
    if (ret < 0) goto cleanup;

    int video_index = av_find_best_stream(video_input, AVMEDIA_TYPE_VIDEO, -1, -1, NULL, 0);
    if (video_index < 0) {
        tt_set_error(err, video_index, "find video stream");
        ret = video_index;
        goto cleanup;
    }
    int audio_index = av_find_best_stream(audio_input, AVMEDIA_TYPE_AUDIO, -1, -1, NULL, 0);
    if (audio_index < 0) {
        tt_set_error(err, audio_index, "find audio stream");
        ret = audio_index;
        goto cleanup;
    }

    ret = avformat_alloc_output_context2(&output, NULL, NULL, output_path);
    if (ret < 0 || !output) {
        tt_set_error(err, ret < 0 ? ret : AVERROR_UNKNOWN, "allocate output context");
        goto cleanup;
    }
    ret = tt_copy_stream(output, video_input->streams[video_index], &video_out, err);
    if (ret < 0) goto cleanup;
    ret = tt_copy_stream(output, audio_input->streams[audio_index], &audio_out, err);
    if (ret < 0) goto cleanup;

    if (!(output->oformat->flags & AVFMT_NOFILE)) {
        ret = avio_open(&output->pb, output_path, AVIO_FLAG_WRITE);
        if (ret < 0) {
            tt_set_error(err, ret, "open output");
            goto cleanup;
        }
    }
    ret = avformat_write_header(output, NULL);
    if (ret < 0) {
        tt_set_error(err, ret, "write header");
        goto cleanup;
    }

    video_packet = av_packet_alloc();
    audio_packet = av_packet_alloc();
    if (!video_packet || !audio_packet) {
        *err = tt_strdup("allocate packet: allocation failed");
        ret = AVERROR(ENOMEM);
        goto cleanup;
    }

    video_ready = tt_read_selected_packet(video_input, video_index, video_packet, err);
    if (video_ready < 0) { ret = video_ready; goto cleanup; }
    audio_ready = tt_read_selected_packet(audio_input, audio_index, audio_packet, err);
    if (audio_ready < 0) { ret = audio_ready; goto cleanup; }

    while (video_ready || audio_ready) {
        int write_video = 0;
        if (video_ready && audio_ready) {
            int64_t vt = tt_packet_time(video_packet, video_input->streams[video_index]->time_base);
            int64_t at = tt_packet_time(audio_packet, audio_input->streams[audio_index]->time_base);
            write_video = vt <= at;
        } else {
            write_video = video_ready;
        }

        if (write_video) {
            ret = tt_write_packet(output, video_packet, video_input->streams[video_index], video_out, err);
            av_packet_unref(video_packet);
            if (ret < 0) goto cleanup;
            video_ready = tt_read_selected_packet(video_input, video_index, video_packet, err);
            if (video_ready < 0) { ret = video_ready; goto cleanup; }
        } else {
            ret = tt_write_packet(output, audio_packet, audio_input->streams[audio_index], audio_out, err);
            av_packet_unref(audio_packet);
            if (ret < 0) goto cleanup;
            audio_ready = tt_read_selected_packet(audio_input, audio_index, audio_packet, err);
            if (audio_ready < 0) { ret = audio_ready; goto cleanup; }
        }
    }

    ret = av_write_trailer(output);
    if (ret < 0) tt_set_error(err, ret, "write trailer");

cleanup:
    if (video_packet) av_packet_free(&video_packet);
    if (audio_packet) av_packet_free(&audio_packet);
    if (output) {
        if (output->pb) avio_closep(&output->pb);
        avformat_free_context(output);
    }
    if (video_input) avformat_close_input(&video_input);
    if (audio_input) avformat_close_input(&audio_input);
    return ret;
}
*/
import "C"

import (
	"context"
	"fmt"
	"unsafe"
)

func RemuxAVFormat(ctx context.Context, videoPath string, audioPath string, outputPath string) error {
	return RemuxAVFormatWithHeaders(ctx, videoPath, "", audioPath, "", outputPath)
}

func RemuxAVFormatWithHeaders(ctx context.Context, videoPath string, videoHeaders string, audioPath string, audioHeaders string, outputPath string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	cVideo := C.CString(videoPath)
	cVideoHeaders := C.CString(videoHeaders)
	cAudio := C.CString(audioPath)
	cAudioHeaders := C.CString(audioHeaders)
	cOutput := C.CString(outputPath)
	defer C.free(unsafe.Pointer(cVideo))
	defer C.free(unsafe.Pointer(cVideoHeaders))
	defer C.free(unsafe.Pointer(cAudio))
	defer C.free(unsafe.Pointer(cAudioHeaders))
	defer C.free(unsafe.Pointer(cOutput))

	var cErr *C.char
	ret := C.tt_remux_avformat(cVideo, cVideoHeaders, cAudio, cAudioHeaders, cOutput, &cErr)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
	}
	if ret < 0 {
		if cErr != nil {
			return fmt.Errorf("libavformat remux failed: %s", C.GoString(cErr))
		}
		return fmt.Errorf("libavformat remux failed: %d", int(ret))
	}
	return ctx.Err()
}
