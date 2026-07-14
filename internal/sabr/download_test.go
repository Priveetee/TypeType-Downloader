package sabr

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestDownloadPreservesOrderAndBoundsConcurrency(t *testing.T) {
	var mu sync.Mutex
	active := 0
	maximum := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer test" {
			t.Errorf("authorization = %q", request.Header.Get("Authorization"))
		}
		if request.URL.Path == "/manifest" {
			writer.Header().Set("Content-Type", "application/dash+xml")
			fmt.Fprint(writer, testManifest(false))
			return
		}
		mu.Lock()
		active++
		if active > maximum {
			maximum = active
		}
		mu.Unlock()
		time.Sleep(30 * time.Millisecond)
		fmt.Fprint(writer, request.URL.Path)
		mu.Lock()
		active--
		mu.Unlock()
	}))
	defer server.Close()

	dir := t.TempDir()
	video := filepath.Join(dir, "video.mp4")
	audio := filepath.Join(dir, "audio.m4a")
	var progress int64
	err := Download(context.Background(), server.Client(), Options{
		ManifestURL: server.URL + "/manifest", Authorization: "Bearer test",
		VideoItag: 137, AudioItag: 140, Workers: 4, WorkDir: dir,
		VideoPath: video, AudioPath: audio,
	}, func(value int64) { progress = value })
	if err != nil {
		t.Fatal(err)
	}
	assertFile(t, video, "/video/init/video/1/video/2")
	assertFile(t, audio, "/audio/init/audio/1/audio/2")
	if maximum < 2 || maximum > 4 {
		t.Fatalf("maximum concurrency = %d, want 2..4", maximum)
	}
	if progress != int64(len("/video/init/video/1/video/2/audio/init/audio/1/audio/2")) {
		t.Fatalf("progress = %d", progress)
	}
}

func TestDownloadRetriesTransientSegmentFailure(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/manifest" {
			fmt.Fprint(writer, testManifest(true))
			return
		}
		if request.URL.Path == "/audio/1" {
			attempts++
			if attempts == 1 {
				http.Error(writer, "not ready", http.StatusNotFound)
				return
			}
		}
		fmt.Fprint(writer, request.URL.Path)
	}))
	defer server.Close()
	dir := t.TempDir()
	output := filepath.Join(dir, "audio.m4a")
	err := Download(context.Background(), server.Client(), Options{
		ManifestURL: server.URL + "/manifest", AudioItag: 140,
		AudioOnly: true, Workers: 1, WorkDir: dir, AudioPath: output,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	assertFile(t, output, "/audio/init/audio/1/audio/2")
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
}

func TestDownloadCancellationRemovesTemporaryFiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/manifest" {
			fmt.Fprint(writer, testManifest(true))
			return
		}
		<-request.Context().Done()
	}))
	defer server.Close()
	dir := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err := Download(ctx, server.Client(), Options{
		ManifestURL: server.URL + "/manifest", AudioItag: 140,
		AudioOnly: true, Workers: 1, WorkDir: dir, AudioPath: filepath.Join(dir, "audio.m4a"),
	}, nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v, want deadline exceeded", err)
	}
	entries, readErr := os.ReadDir(dir)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("temporary files remain: %v", entries)
	}
}

func testManifest(audioOnly bool) string {
	audio := `<AdaptationSet mimeType="audio/mp4"><Representation><SegmentList><Initialization sourceURL="/audio/init"/><SegmentURL media="/audio/1"/><SegmentURL media="/audio/2"/></SegmentList></Representation></AdaptationSet>`
	if audioOnly {
		return `<MPD><Period>` + audio + `</Period></MPD>`
	}
	video := `<AdaptationSet mimeType="video/mp4"><Representation><SegmentList><Initialization sourceURL="/video/init"/><SegmentURL media="/video/1"/><SegmentURL media="/video/2"/></SegmentList></Representation></AdaptationSet>`
	return `<MPD><Period>` + video + audio + `</Period></MPD>`
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
