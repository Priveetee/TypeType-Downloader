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

func Start(ctx context.Context, dataDir string) {
	go func() {
		clean(dataDir, time.Hour)
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				clean(dataDir, time.Hour)
			}
		}
	}()
}

func clean(dataDir string, maxAge time.Duration) {
	jobsDir := filepath.Join(dataDir, "jobs")
	cutoff := time.Now().Add(-maxAge)
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
