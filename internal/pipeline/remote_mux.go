package pipeline

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"typetype-downloader-go/internal/artifact"
	"typetype-downloader-go/internal/downloader"
	"typetype-downloader-go/internal/job"
	"typetype-downloader-go/internal/mux"
	"typetype-downloader-go/internal/selector"
)

func (r *Runner) runRemoteMux(ctx context.Context, id string, title string, selection *selector.Selection) error {
	started := time.Now()
	paths := artifact.Build(r.cfg.DataDir, id, title, selection.Container)
	preserveOutput := false
	defer func() { cleanupWork(paths, preserveOutput) }()
	r.store.Resolve(id, title, resolvedOutput(selection, paths.Name))
	if err := os.MkdirAll(paths.WorkDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(paths.Output), 0o755); err != nil {
		return err
	}
	muxStarted := time.Now()
	r.store.Progress(id, job.Progress{Stage: "mux"})
	if err := retry(ctx, 2, "mux", func() error {
		_ = os.Remove(paths.Output)
		return mergeRemote(ctx, r.cfg.Muxer, selection.Video.URL, selection.Audio.URL, paths.Output)
	}); err != nil {
		return err
	}
	muxMs := time.Since(muxStarted).Milliseconds()
	var saved artifact.Saved
	if err := retry(ctx, 3, "artifact upload", func() error {
		var saveErr error
		saved, saveErr = r.storage.Save(ctx, paths.Output, paths.Key)
		return saveErr
	}); err != nil {
		return err
	}
	expiresAt := saved.Expires
	var expires *time.Time
	if !expiresAt.IsZero() {
		expires = &expiresAt
	}
	r.store.Done(id, saved.Location, saved.Backend, expires, 0, muxMs)
	preserveOutput = saved.Backend == "local"
	slog.Info("remote mux job completed", "id", id, "ms", time.Since(started).Milliseconds())
	return nil
}

func mergeRemote(ctx context.Context, muxer string, videoURL string, audioURL string, outputPath string) error {
	if muxer != "avformat" {
		return fmt.Errorf("remote mux requires avformat muxer")
	}
	videoPath, err := mediaPath(videoURL)
	if err != nil {
		return err
	}
	audioPath, err := mediaPath(audioURL)
	if err != nil {
		return err
	}
	return mux.RemuxAVFormatWithHeaders(ctx, videoPath, headerLines(videoURL), audioPath, headerLines(audioURL), outputPath)
}

func usesRemoteMux(selection *selector.Selection) bool {
	return isHLS(selection.Video.URL) || isHLS(selection.Audio.URL)
}

func isHLS(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return strings.HasSuffix(strings.ToLower(parsed.Path), ".m3u8")
}

func mediaPath(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	parsed.Fragment = ""
	return parsed.String(), nil
}

func headerLines(rawURL string) string {
	headers := downloader.MediaHeaders(rawURL, false)
	var builder strings.Builder
	for name, values := range headers {
		for _, value := range values {
			builder.WriteString(name)
			builder.WriteString(": ")
			builder.WriteString(value)
			builder.WriteString("\r\n")
		}
	}
	return builder.String()
}
