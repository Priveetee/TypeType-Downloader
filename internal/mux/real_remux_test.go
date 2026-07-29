package mux

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRealRemux(t *testing.T) {
	videoPath := os.Getenv("TYPETYPE_REMUX_VIDEO")
	audioPath := os.Getenv("TYPETYPE_REMUX_AUDIO")
	if videoPath == "" || audioPath == "" {
		t.Skip("set TYPETYPE_REMUX_VIDEO and TYPETYPE_REMUX_AUDIO to enable the real remux test")
	}
	outputPath := os.Getenv("TYPETYPE_REMUX_OUTPUT")
	if outputPath == "" {
		outputPath = filepath.Join(t.TempDir(), "output.mp4")
	}
	started := time.Now()
	if err := RemuxAVFormat(context.Background(), videoPath, audioPath, outputPath); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("duration=%s bytes=%d", time.Since(started), info.Size())
}
