package cleanup

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func oldArtifact(t *testing.T, dataDir string) string {
	t.Helper()
	path := filepath.Join(dataDir, "artifacts", "old.mp4")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("artifact"), 0o644); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCleanRemovesUploadedS3Artifacts(t *testing.T) {
	dataDir := t.TempDir()
	path := oldArtifact(t, dataDir)

	clean(dataDir, "s3", time.Hour)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("artifact still exists: %v", err)
	}
}

func TestCleanKeepsLocalArtifacts(t *testing.T) {
	dataDir := t.TempDir()
	path := oldArtifact(t, dataDir)

	clean(dataDir, "local", time.Hour)

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("artifact missing: %v", err)
	}
}
