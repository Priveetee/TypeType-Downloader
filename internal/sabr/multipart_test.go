package sabr

import (
	"context"
	"net/http"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func TestDownloadBoundsConcurrentSABRParts(t *testing.T) {
	var active atomic.Int64
	var peak atomic.Int64
	var requests atomic.Int64
	server := downloadTestServer(t, func(writer http.ResponseWriter, request *http.Request) {
		current := active.Add(1)
		defer active.Add(-1)
		requests.Add(1)
		for {
			observed := peak.Load()
			if current <= observed || peak.CompareAndSwap(observed, current) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
		part, err := strconv.Atoi(request.URL.Query().Get("part"))
		if err != nil {
			t.Error(err)
			return
		}
		writeTestFrame(t, writer, frameInitialization, 140, 0, []byte("init"))
		writeTestFrame(t, writer, frameMedia, 140, part+1, []byte("media"))
		writeTestFrame(t, writer, frameComplete, 0, 0, nil)
	})
	defer server.Close()

	output := filepath.Join(t.TempDir(), "audio.m4a")
	err := Download(context.Background(), server.Client(), Options{
		ManifestURL: server.URL,
		AudioItag:   140,
		AudioOnly:   true,
		AudioPath:   output,
		Parts:       maxDownloadParts,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if requests.Load() != maxDownloadParts {
		t.Fatalf("requests = %d, want %d", requests.Load(), maxDownloadParts)
	}
	if peak.Load() != maxConcurrentPartDownloads {
		t.Fatalf("peak concurrency = %d, want %d", peak.Load(), maxConcurrentPartDownloads)
	}
}
