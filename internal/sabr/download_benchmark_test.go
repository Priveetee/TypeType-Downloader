package sabr

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func BenchmarkDownloadVideoFifteenMinutes(b *testing.B) {
	benchmarkDownload(b, false, 180, 1)
}

func BenchmarkDownloadAudioThirtyMinutes(b *testing.B) {
	benchmarkDownload(b, true, 360, 1)
}

func BenchmarkDownloadVideoFifteenMinutesMultipart(b *testing.B) {
	benchmarkDownload(b, false, 180, 4)
}

func BenchmarkDownloadAudioThirtyMinutesMultipart(b *testing.B) {
	benchmarkDownload(b, true, 360, 4)
}

func BenchmarkDownloadVideoTenHours(b *testing.B) {
	benchmarkDownload(b, false, 7200, 12)
}

func BenchmarkDownloadAudioTenHours(b *testing.B) {
	benchmarkDownload(b, true, 7200, 4)
}

func benchmarkDownload(b *testing.B, audioOnly bool, segments int, parts int) {
	streams := make([][]byte, parts)
	totalBytes := int64(0)
	for part := range parts {
		streams[part] = benchmarkStreamPart(b, audioOnly, segments, part, parts)
		totalBytes += int64(len(streams[part]))
	}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", downloadMediaType)
		part, _ := strconv.Atoi(request.URL.Query().Get("part"))
		_, _ = writer.Write(streams[part])
	}))
	b.Cleanup(server.Close)
	b.ReportAllocs()
	b.SetBytes(totalBytes)
	b.ResetTimer()
	for range b.N {
		dir, err := os.MkdirTemp("", "sabr-stream-benchmark-")
		if err != nil {
			b.Fatal(err)
		}
		options := Options{
			ManifestURL: server.URL,
			VideoItag:   137,
			AudioItag:   140,
			AudioOnly:   audioOnly,
			VideoPath:   filepath.Join(dir, "video.mp4"),
			AudioPath:   filepath.Join(dir, "audio.m4a"),
			Parts:       parts,
		}
		if err := Download(context.Background(), server.Client(), options, nil); err != nil {
			_ = os.RemoveAll(dir)
			b.Fatal(err)
		}
		_ = os.RemoveAll(dir)
	}
}

func benchmarkStreamPart(t testing.TB, audioOnly bool, segments int, part int, parts int) []byte {
	t.Helper()
	payload := bytes.Repeat([]byte{0x5a}, 16<<10)
	var stream bytes.Buffer
	stream.Write(downloadMagic)
	writeTestFrame(t, &stream, frameInitialization, 140, 0, []byte("audio-init"))
	if !audioOnly {
		writeTestFrame(t, &stream, frameInitialization, 137, 0, []byte("video-init"))
	}
	start := segments*part/parts + 1
	end := segments * (part + 1) / parts
	for sequence := start; sequence <= end; sequence++ {
		writeTestFrame(t, &stream, frameMedia, 140, sequence, payload)
		if !audioOnly {
			writeTestFrame(t, &stream, frameMedia, 137, sequence, payload)
		}
	}
	writeTestFrame(t, &stream, frameComplete, 0, 0, nil)
	return stream.Bytes()
}
