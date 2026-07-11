package pipeline

import (
	"context"
	"errors"
	"testing"
)

func TestFailureCodeUsesCancelledContextForWrappedProcessError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if code := failureCode(ctx, errors.New("ffmpeg failed: signal: killed")); code != "cancelled" {
		t.Fatalf("expected cancelled, got %q", code)
	}
}

func TestFailureCodeKeepsDownloadFailure(t *testing.T) {
	if code := failureCode(context.Background(), errors.New("download failed")); code != "download_failed" {
		t.Fatalf("expected download_failed, got %q", code)
	}
}
