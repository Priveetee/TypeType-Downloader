package sabr

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestNetworkDownload(t *testing.T) {
	manifestURL := os.Getenv("TYPETYPE_SABR_MANIFEST_URL")
	if manifestURL == "" {
		t.Skip("set TYPETYPE_SABR_MANIFEST_URL to enable the network test")
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DisableCompression = true
	transport.ForceAttemptHTTP2 = true
	transport.MaxIdleConns = 8
	transport.MaxIdleConnsPerHost = 8
	transport.MaxConnsPerHost = 8
	client := &http.Client{Transport: transport}
	t.Cleanup(transport.CloseIdleConnections)

	dir := os.Getenv("TYPETYPE_SABR_OUTPUT_DIR")
	if dir == "" {
		dir = t.TempDir()
	} else if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	err := Download(context.Background(), client, Options{
		ManifestURL:   manifestURL,
		Authorization: os.Getenv("TYPETYPE_SABR_AUTHORIZATION"),
		VideoItag:     networkInt("TYPETYPE_SABR_VIDEO_ITAG", 137),
		AudioItag:     networkInt("TYPETYPE_SABR_AUDIO_ITAG", 140),
		AudioTrackID:  os.Getenv("TYPETYPE_SABR_AUDIO_TRACK_ID"),
		AudioOnly:     os.Getenv("TYPETYPE_SABR_AUDIO_ONLY") == "1",
		VideoPath:     filepath.Join(dir, "video.mp4"),
		AudioPath:     filepath.Join(dir, "audio.m4a"),
		Parts:         networkInt("TYPETYPE_SABR_PARTS", 1),
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("duration=%s bytes=%d", time.Since(started), outputBytes(t, dir))
	for _, name := range []string{"video.mp4", "audio.m4a"} {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			t.Logf("%s sha256=%s", name, fileSHA256(t, path))
		}
	}
}

func fileSHA256(t *testing.T, path string) string {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func networkInt(name string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(name))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func outputBytes(t *testing.T, dir string) int64 {
	t.Helper()
	var total int64
	for _, name := range []string{"video.mp4", "audio.m4a"} {
		info, err := os.Stat(filepath.Join(dir, name))
		if err == nil {
			total += info.Size()
		} else if !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	return total
}
