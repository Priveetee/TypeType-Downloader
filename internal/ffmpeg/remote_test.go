package ffmpeg

import (
	"slices"
	"testing"
)

func TestRemoteInputAddsBiliBiliReferer(t *testing.T) {
	args := remoteInput("https://upos-hz-mirrorakam.akamaized.net/video.m4s")
	if !slices.Contains(args, "Referer: https://www.bilibili.com\r\n") {
		t.Fatalf("args = %v", args)
	}
}

func TestRemoteInputLeavesNicoHeadersGeneric(t *testing.T) {
	args := remoteInput("http://server:8080/proxy/nicovideo?url=video.m3u8")
	if slices.Contains(args, "-headers") {
		t.Fatalf("args = %v", args)
	}
	if !slices.Contains(args, "-allowed_segment_extensions") || !slices.Contains(args, "ALL") {
		t.Fatalf("args = %v", args)
	}
}
