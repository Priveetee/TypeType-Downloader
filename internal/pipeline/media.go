package pipeline

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"typetype-downloader-go/internal/ffmpeg"
	"typetype-downloader-go/internal/job"
	"typetype-downloader-go/internal/mux"
	"typetype-downloader-go/internal/selector"
)

func newHTTPClient(workers int, http2 bool) *http.Client {
	maxConns := workers * 8
	if maxConns < 32 {
		maxConns = 32
	}
	transport := &http.Transport{
		DisableCompression:  true,
		ForceAttemptHTTP2:   http2,
		MaxIdleConns:        512,
		MaxIdleConnsPerHost: maxConns,
		MaxConnsPerHost:     maxConns,
		IdleConnTimeout:     90 * time.Second,
	}
	return &http.Client{Transport: transport}
}

func merge(ctx context.Context, muxer string, videoPath string, audioPath string, outputPath string) error {
	if muxer == "avformat" {
		return mux.RemuxAVFormat(ctx, videoPath, audioPath, outputPath)
	}
	if muxer == "ffmpeg" {
		return ffmpeg.MergeCopy(ctx, videoPath, audioPath, outputPath)
	}
	return fmt.Errorf("unsupported muxer %q", muxer)
}

func runParallel(tasks ...func() error) error {
	var wg sync.WaitGroup
	errs := make(chan error, len(tasks))
	for _, task := range tasks {
		wg.Add(1)
		go func(task func() error) {
			defer wg.Done()
			if err := task(); err != nil {
				errs <- err
			}
		}(task)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		return err
	}
	return nil
}

func resolvedOutput(selection *selector.Selection, fileName string) job.ResolvedOutput {
	return job.ResolvedOutput{
		VideoItag:  strconv.Itoa(selection.Video.Itag),
		AudioItag:  strconv.Itoa(selection.Audio.Itag),
		Height:     selection.Video.Height,
		FPS:        selection.Video.FPS,
		VideoCodec: stringValue(selection.Video.Codec),
		AudioCodec: stringValue(selection.Audio.Codec),
		Container:  selection.Container,
		FileName:   fileName,
	}
}

func audioResolvedOutput(selection *selector.AudioSelection, fileName string) job.ResolvedOutput {
	return job.ResolvedOutput{
		AudioItag:  strconv.Itoa(selection.Audio.Itag),
		AudioCodec: stringValue(selection.Audio.Codec),
		Container:  selection.Container,
		FileName:   fileName,
	}
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
