package sabr

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestDownloadWritesOrderedTracksAtomically(t *testing.T) {
	server := downloadTestServer(t, func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer test" {
			t.Errorf("authorization = %q", request.Header.Get("Authorization"))
		}
		if request.URL.Path != "/sabr/download/video" ||
			request.URL.Query().Get("audioItag") != "140" ||
			request.URL.Query().Get("videoItag") != "137" ||
			request.URL.Query().Get("part") != "0" ||
			request.URL.Query().Get("parts") != "1" {
			t.Errorf("query = %q", request.URL.RawQuery)
		}
		writeTestFrame(t, writer, frameInitialization, 140, 0, []byte("ai"))
		writeTestFrame(t, writer, frameInitialization, 137, 0, []byte("vi"))
		writeTestFrame(t, writer, frameMedia, 137, 1, []byte("v1"))
		writeTestFrame(t, writer, frameMedia, 140, 1, []byte("a1"))
		writeTestFrame(t, writer, frameMedia, 140, 2, []byte("a2"))
		writeTestFrame(t, writer, frameMedia, 137, 2, []byte("v2"))
		writeTestFrame(t, writer, frameComplete, 0, 0, nil)
	})
	defer server.Close()

	dir := t.TempDir()
	video := filepath.Join(dir, "video.mp4")
	audio := filepath.Join(dir, "audio.m4a")
	var progress int64
	err := Download(context.Background(), server.Client(), Options{
		ManifestURL: server.URL + "/sabr/manifest/video", Authorization: "Bearer test",
		VideoItag: 137, AudioItag: 140, VideoPath: video, AudioPath: audio,
	}, func(value int64) { progress = value })
	if err != nil {
		t.Fatal(err)
	}
	assertFile(t, video, "viv1v2")
	assertFile(t, audio, "aia1a2")
	if progress != 12 {
		t.Fatalf("progress = %d, want 12", progress)
	}
}

func TestDownloadAssemblesMultipartTracksWithoutDuplicateInitialization(t *testing.T) {
	server := downloadTestServer(t, func(writer http.ResponseWriter, request *http.Request) {
		part := request.URL.Query().Get("part")
		switch part {
		case "0":
			writeTestFrame(t, writer, frameInitialization, 140, 0, []byte("ai"))
			writeTestFrame(t, writer, frameInitialization, 137, 0, []byte("vi"))
			writeTestFrame(t, writer, frameMedia, 140, 1, []byte("a1"))
			writeTestFrame(t, writer, frameMedia, 137, 1, []byte("v1"))
		case "1":
			writeTestFrame(t, writer, frameInitialization, 140, 0, []byte("duplicate-ai"))
			writeTestFrame(t, writer, frameInitialization, 137, 0, []byte("duplicate-vi"))
			writeTestFrame(t, writer, frameMedia, 140, 7, []byte("a7"))
			writeTestFrame(t, writer, frameMedia, 137, 4, []byte("v4"))
		default:
			t.Fatalf("unexpected part %q", part)
		}
		writeTestFrame(t, writer, frameComplete, 0, 0, nil)
	})
	defer server.Close()

	dir := t.TempDir()
	video := filepath.Join(dir, "video.mp4")
	audio := filepath.Join(dir, "audio.m4a")
	var progress int64
	err := Download(context.Background(), server.Client(), Options{
		ManifestURL: server.URL, VideoItag: 137, AudioItag: 140,
		VideoPath: video, AudioPath: audio, Parts: 2,
	}, func(value int64) { progress = value })
	if err != nil {
		t.Fatal(err)
	}
	assertFile(t, video, "viv1v4")
	assertFile(t, audio, "aia1a7")
	if progress != 12 {
		t.Fatalf("progress = %d, want 12", progress)
	}
	assertNoFile(t, partPath(video, 0))
	assertNoFile(t, partPath(video, 1))
	assertNoFile(t, partPath(audio, 0))
	assertNoFile(t, partPath(audio, 1))
}

func TestDownloadRetriesTruncatedStream(t *testing.T) {
	var attempts atomic.Int64
	server := downloadTestServer(t, func(writer http.ResponseWriter, _ *http.Request) {
		if attempts.Add(1) == 1 {
			writeTestFrame(t, writer, frameInitialization, 140, 0, []byte("init"))
			return
		}
		writeTestFrame(t, writer, frameInitialization, 140, 0, []byte("init"))
		writeTestFrame(t, writer, frameMedia, 140, 1, []byte("media"))
		writeTestFrame(t, writer, frameComplete, 0, 0, nil)
	})
	defer server.Close()

	output := filepath.Join(t.TempDir(), "audio.m4a")
	err := Download(context.Background(), server.Client(), Options{
		ManifestURL: server.URL, AudioItag: 140, AudioOnly: true, AudioPath: output,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	assertFile(t, output, "initmedia")
	if attempts.Load() != 2 {
		t.Fatalf("attempts = %d, want 2", attempts.Load())
	}
}

func TestReporterKeepsRetryProgressMonotonic(t *testing.T) {
	var updates []int64
	reporter := newReporter(func(value int64) {
		updates = append(updates, value)
	})

	reporter.beginAttempt()
	reporter.reportMu.Lock()
	reporter.last = time.Now().Add(-time.Second)
	reporter.reportMu.Unlock()
	reporter.add(10)

	reporter.beginAttempt()
	reporter.reportMu.Lock()
	reporter.last = time.Now().Add(-time.Second)
	reporter.reportMu.Unlock()
	reporter.add(5)
	reporter.add(10)
	reporter.finish()

	if len(updates) != 2 || updates[0] != 10 || updates[1] != 15 {
		t.Fatalf("updates = %v, want [10 15]", updates)
	}
}

func TestDownloadRejectsOutOfOrderMediaAndPreservesTarget(t *testing.T) {
	server := downloadTestServer(t, func(writer http.ResponseWriter, _ *http.Request) {
		writeTestFrame(t, writer, frameInitialization, 140, 0, []byte("init"))
		writeTestFrame(t, writer, frameMedia, 140, 2, []byte("wrong"))
	})
	defer server.Close()

	dir := t.TempDir()
	output := filepath.Join(dir, "audio.m4a")
	if err := os.WriteFile(output, []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := Download(context.Background(), server.Client(), Options{
		ManifestURL: server.URL, AudioItag: 140, AudioOnly: true, AudioPath: output,
	}, nil)
	if err == nil {
		t.Fatal("expected out-of-order stream failure")
	}
	assertFile(t, output, "existing")
	assertNoFile(t, output+".download")
}

func TestDownloadCancellationRemovesTemporaryFile(t *testing.T) {
	server := downloadTestServer(t, func(writer http.ResponseWriter, request *http.Request) {
		writeTestFrame(t, writer, frameInitialization, 140, 0, []byte("init"))
		writer.(http.Flusher).Flush()
		<-request.Context().Done()
	})
	defer server.Close()

	output := filepath.Join(t.TempDir(), "audio.m4a")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err := Download(ctx, server.Client(), Options{
		ManifestURL: server.URL, AudioItag: 140, AudioOnly: true, AudioPath: output,
	}, nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v, want deadline exceeded", err)
	}
	assertNoFile(t, output)
	assertNoFile(t, output+".download")
}

func TestDownloadRejectsOversizedFrame(t *testing.T) {
	server := downloadTestServer(t, func(writer http.ResponseWriter, _ *http.Request) {
		writeTestHeader(t, writer, frameInitialization, 140, 0, maxFrameBytes+1)
	})
	defer server.Close()
	output := filepath.Join(t.TempDir(), "audio.m4a")
	err := Download(context.Background(), server.Client(), Options{
		ManifestURL: server.URL, AudioItag: 140, AudioOnly: true, AudioPath: output,
	}, nil)
	if err == nil {
		t.Fatal("expected oversized frame failure")
	}
	assertNoFile(t, output)
	assertNoFile(t, output+".download")
}

func TestDownloadPartCountBoundsAudioFanout(t *testing.T) {
	if got := downloadPartCount(Options{AudioOnly: true, ExpectedBytes: 34 << 20}); got != 4 {
		t.Fatalf("audio parts = %d, want 4", got)
	}
	if got := downloadPartCount(Options{ExpectedBytes: 34 << 20}); got != 4 {
		t.Fatalf("video parts = %d, want 4", got)
	}
	if got := downloadPartCount(Options{ExpectedBytes: 416 << 20}); got != 12 {
		t.Fatalf("large video parts = %d, want 12", got)
	}
	if got := downloadPartCount(Options{AudioOnly: true, ExpectedBytes: 34 << 20, Parts: 6}); got != 6 {
		t.Fatalf("explicit audio parts = %d, want 6", got)
	}
	if got := downloadPartCount(Options{ExpectedBytes: 1 << 30, Parts: 20}); got != 12 {
		t.Fatalf("capped parts = %d, want 12", got)
	}
}

func downloadTestServer(
	t *testing.T,
	handler func(http.ResponseWriter, *http.Request),
) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", downloadMediaType)
		if _, err := writer.Write(downloadMagic); err != nil {
			return
		}
		handler(writer, request)
	}))
}

func writeTestFrame(
	t testing.TB,
	writer io.Writer,
	kind byte,
	itag int,
	sequence int,
	payload []byte,
) {
	t.Helper()
	writeTestHeader(t, writer, kind, itag, sequence, int64(len(payload)))
	if _, err := writer.Write(payload); err != nil {
		t.Fatal(err)
	}
}

func writeTestHeader(
	t testing.TB,
	writer io.Writer,
	kind byte,
	itag int,
	sequence int,
	length int64,
) {
	t.Helper()
	var header [frameHeaderSize]byte
	header[0] = kind
	binary.BigEndian.PutUint32(header[1:5], uint32(itag))
	binary.BigEndian.PutUint32(header[5:9], uint32(sequence))
	binary.BigEndian.PutUint64(header[9:17], uint64(length))
	if _, err := writer.Write(header[:]); err != nil {
		t.Fatal(err)
	}
}

func assertFile(t *testing.T, path string, expected string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != expected {
		t.Fatalf("%s = %q, want %q", path, content, expected)
	}
}

func assertNoFile(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("%s exists or stat failed: %v", path, err)
	}
}
