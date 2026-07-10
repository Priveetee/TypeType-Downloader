package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"typetype-downloader-go/internal/artifact"
)

func TestCleanupWorkPreservesCompletedLocalOutput(t *testing.T) {
	dir := t.TempDir()
	paths := artifact.Paths{
		Video:  filepath.Join(dir, "video.mp4"),
		Audio:  filepath.Join(dir, "audio.m4a"),
		Output: filepath.Join(dir, "output.mp4"),
	}
	for _, path := range []string{paths.Video, paths.Audio, paths.Output} {
		if err := os.WriteFile(path, []byte("media"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	cleanupWork(paths, true)

	if _, err := os.Stat(paths.Output); err != nil {
		t.Fatalf("output missing: %v", err)
	}
	for _, path := range []string{paths.Video, paths.Audio} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("work file still exists: %s", path)
		}
	}
}

func TestCleanupWorkRemovesFailedOutput(t *testing.T) {
	dir := t.TempDir()
	paths := artifact.Paths{Output: filepath.Join(dir, "output.mp4")}
	if err := os.WriteFile(paths.Output, []byte("media"), 0o644); err != nil {
		t.Fatal(err)
	}

	cleanupWork(paths, false)

	if _, err := os.Stat(paths.Output); !os.IsNotExist(err) {
		t.Fatalf("output still exists: %v", err)
	}
}
