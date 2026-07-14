package cleanup

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func Start(ctx context.Context, dataDir string, storageBackend string) {
	go func() {
		clean(dataDir, storageBackend, time.Hour)
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				clean(dataDir, storageBackend, time.Hour)
			}
		}
	}()
}

func clean(dataDir string, storageBackend string, maxAge time.Duration) {
	cutoff := time.Now().Add(-maxAge)
	cleanJobs(filepath.Join(dataDir, "jobs"), cutoff)
	if storageBackend == "s3" {
		cleanArtifacts(filepath.Join(dataDir, "artifacts"), cutoff)
	}
}

func cleanJobs(jobsDir string, cutoff time.Time) {
	_ = filepath.WalkDir(jobsDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil || info.ModTime().After(cutoff) {
			return nil
		}
		if strings.HasSuffix(path, ".part") || strings.Contains(path, string(filepath.Separator)+"video.") || strings.Contains(path, string(filepath.Separator)+"audio.") {
			if err := os.Remove(path); err != nil {
				slog.Warn("cleanup failed", "path", path, "error", err)
			}
		}
		return nil
	})
}

func cleanArtifacts(artifactsDir string, cutoff time.Time) {
	_ = filepath.WalkDir(artifactsDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err == nil && info.ModTime().Before(cutoff) {
			remove(path)
		}
		return nil
	})
}

func remove(path string) {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		slog.Warn("cleanup failed", "path", path, "error", err)
	}
}
