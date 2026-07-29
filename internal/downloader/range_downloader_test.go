package downloader

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
)

func TestDownloadFileWritesConcurrentRangesExactly(t *testing.T) {
	data := repeatedPayload(3*64*1024 + 713)
	for _, mode := range []string{"header", "query"} {
		t.Run(mode, func(t *testing.T) {
			var requests atomic.Int64
			server := httptest.NewServer(rangeServer(data, &requests, nil))
			defer server.Close()

			output := filepath.Join(t.TempDir(), "media.bin")
			var updates []Progress
			err := DownloadFile(t.Context(), server.Client(), Source{
				Name: "video",
				URL:  server.URL + "/media#cookie=ignored",
				Size: int64(len(data)),
			}, output, Options{
				ChunkSize:     64 * 1024,
				Workers:       32,
				Retries:       1,
				BufferSize:    16 * 1024,
				RangeMode:     mode,
				ProgressBytes: 64 * 1024,
			}, func(progress Progress) {
				updates = append(updates, progress)
			})
			if err != nil {
				t.Fatalf("DownloadFile() error = %v", err)
			}

			got, err := os.ReadFile(output)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, data) {
				t.Fatal("downloaded data differs from source")
			}
			if got := requests.Load(); got != 4 {
				t.Fatalf("requests = %d, want 4", got)
			}
			assertProgressMonotonic(t, updates, int64(len(data)))
		})
	}
}

func TestDownloadFileRetriesTruncatedRange(t *testing.T) {
	data := repeatedPayload(80 * 1024)
	var attempts atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		start, end, err := parseRequestedRange(request)
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		part := data[start : end+1]
		response.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(data)))
		response.Header().Set("Content-Length", strconv.Itoa(len(part)))
		response.WriteHeader(http.StatusPartialContent)
		if attempts.Add(1) == 1 {
			_, _ = response.Write(part[:len(part)/2])
			return
		}
		_, _ = response.Write(part)
	}))
	defer server.Close()

	output := filepath.Join(t.TempDir(), "media.bin")
	err := DownloadFile(t.Context(), server.Client(), Source{
		Name: "audio",
		URL:  server.URL,
		Size: int64(len(data)),
	}, output, Options{
		ChunkSize:  int64(len(data)),
		Workers:    1,
		Retries:    2,
		BufferSize: 32 * 1024,
		RangeMode:  "header",
	}, nil)
	if err != nil {
		t.Fatalf("DownloadFile() error = %v", err)
	}
	if got := attempts.Load(); got != 2 {
		t.Fatalf("attempts = %d, want 2", got)
	}
	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Fatal("downloaded data differs after retry")
	}
}

func TestDownloadFilePreservesOutputAndRemovesPartialFileOnFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		http.Error(response, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	dir := t.TempDir()
	output := filepath.Join(dir, "media.bin")
	if err := os.WriteFile(output, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := DownloadFile(t.Context(), server.Client(), Source{
		Name: "video",
		URL:  server.URL,
		Size: 1024,
	}, output, Options{Retries: 1}, nil)
	if err == nil {
		t.Fatal("DownloadFile() error = nil, want failure")
	}
	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "existing" {
		t.Fatalf("output = %q, want existing data", got)
	}
	if _, err := os.Stat(output + ".part"); !os.IsNotExist(err) {
		t.Fatalf("partial file still exists: %v", err)
	}
}

func TestDownloadFileRejectsChunkOverflow(t *testing.T) {
	data := repeatedPayload(32 * 1024)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		flusher := response.(http.Flusher)
		response.WriteHeader(http.StatusOK)
		_, _ = response.Write(data)
		flusher.Flush()
		_, _ = response.Write([]byte{1})
	}))
	defer server.Close()

	output := filepath.Join(t.TempDir(), "media.bin")
	err := DownloadFile(t.Context(), server.Client(), Source{
		Name: "audio",
		URL:  server.URL,
		Size: int64(len(data)),
	}, output, Options{
		ChunkSize: int64(len(data)),
		Workers:   1,
		Retries:   1,
		RangeMode: "query",
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "overflow") {
		t.Fatalf("DownloadFile() error = %v, want overflow", err)
	}
}

func BenchmarkDownloadFile(b *testing.B) {
	data := repeatedPayload(16 << 20)
	server := httptest.NewServer(rangeServer(data, nil, nil))
	defer server.Close()
	output := filepath.Join(b.TempDir(), "media.bin")
	options := Options{
		ChunkSize:  1 << 20,
		Workers:    8,
		Retries:    1,
		BufferSize: defaultCopyBufferSize,
		RangeMode:  "header",
	}

	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if err := DownloadFile(b.Context(), server.Client(), Source{
			Name: "video",
			URL:  server.URL,
			Size: int64(len(data)),
		}, output, options, nil); err != nil {
			b.Fatal(err)
		}
	}
}

func rangeServer(data []byte, requests *atomic.Int64, truncated *atomic.Bool) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		if requests != nil {
			requests.Add(1)
		}
		start, end, err := parseRequestedRange(request)
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		part := data[start : end+1]
		response.Header().Set("Content-Length", strconv.Itoa(len(part)))
		if request.Header.Get("Range") != "" {
			response.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(data)))
			response.WriteHeader(http.StatusPartialContent)
		}
		if truncated != nil && truncated.CompareAndSwap(false, true) {
			part = part[:len(part)/2]
		}
		for len(part) > 0 {
			size := min(3*1024, len(part))
			_, _ = response.Write(part[:size])
			part = part[size:]
		}
	}
}

func parseRequestedRange(request *http.Request) (int, int, error) {
	value := request.URL.Query().Get("range")
	if value == "" {
		value = strings.TrimPrefix(request.Header.Get("Range"), "bytes=")
	}
	startText, endText, ok := strings.Cut(value, "-")
	if !ok {
		return 0, 0, fmt.Errorf("invalid range %q", value)
	}
	start, startErr := strconv.Atoi(startText)
	end, endErr := strconv.Atoi(endText)
	if startErr != nil || endErr != nil || start < 0 || end < start {
		return 0, 0, fmt.Errorf("invalid range %q", value)
	}
	return start, end, nil
}

func repeatedPayload(size int) []byte {
	pattern := []byte("typetype-downloader-performance")
	return bytes.Repeat(pattern, (size+len(pattern)-1)/len(pattern))[:size]
}

func assertProgressMonotonic(t *testing.T, updates []Progress, total int64) {
	t.Helper()
	var previous int64
	for _, update := range updates {
		if update.Downloaded <= previous {
			t.Fatalf("progress moved from %d to %d", previous, update.Downloaded)
		}
		previous = update.Downloaded
	}
	if previous != total {
		t.Fatalf("final progress = %d, want %d", previous, total)
	}
}
