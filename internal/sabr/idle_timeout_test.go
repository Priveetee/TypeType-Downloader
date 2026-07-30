package sabr

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"testing"
	"time"
)

func TestDownloadRetriesIdleStreamThenFails(t *testing.T) {
	server := downloadTestServer(t, func(writer http.ResponseWriter, request *http.Request) {
		writeTestFrame(t, writer, frameInitialization, 140, 0, []byte("init"))
		writer.(http.Flusher).Flush()
		<-request.Context().Done()
	})
	defer server.Close()

	output := filepath.Join(t.TempDir(), "audio.m4a")
	started := time.Now()
	err := Download(context.Background(), server.Client(), Options{
		ManifestURL: server.URL,
		AudioItag:   140,
		AudioOnly:   true,
		AudioPath:   output,
		IdleTimeout: 30 * time.Millisecond,
	}, nil)
	if !errors.Is(err, errDownloadIdleTimeout) {
		t.Fatalf("error = %v, want idle timeout", err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("idle download took %s", elapsed)
	}
	assertNoFile(t, output)
	assertNoFile(t, output+".download")
}

func TestDownloadIdleTimeoutSlidesWithStreamActivity(t *testing.T) {
	server := downloadTestServer(t, func(writer http.ResponseWriter, _ *http.Request) {
		writeTestFrame(t, writer, frameInitialization, 140, 0, []byte("init"))
		for sequence := 1; sequence <= 4; sequence++ {
			time.Sleep(30 * time.Millisecond)
			writeTestFrame(t, writer, frameMedia, 140, sequence, []byte("media"))
			writer.(http.Flusher).Flush()
		}
		writeTestFrame(t, writer, frameComplete, 0, 0, nil)
	})
	defer server.Close()

	output := filepath.Join(t.TempDir(), "audio.m4a")
	err := Download(context.Background(), server.Client(), Options{
		ManifestURL: server.URL,
		AudioItag:   140,
		AudioOnly:   true,
		AudioPath:   output,
		IdleTimeout: 100 * time.Millisecond,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	assertFile(t, output, "initmediamediamediamedia")
}
