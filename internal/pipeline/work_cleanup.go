package pipeline

import (
	"log/slog"
	"os"

	"typetype-downloader-go/internal/artifact"
)

func cleanupWork(paths artifact.Paths, preserveOutput bool) {
	removePath(paths.Video)
	removePath(paths.Audio)
	if !preserveOutput {
		removePath(paths.Output)
	}
}

func removePath(path string) {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		slog.Warn("work cleanup failed", "path", path, "error", err)
	}
}
