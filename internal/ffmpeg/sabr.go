package ffmpeg

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
)

type SABROptions struct {
	ManifestURL   string
	Authorization string
	VideoItag     int
	AudioItag     int
	AudioTrackID  string
	AudioOnly     bool
}

type ProgressFunc func(downloadedBytes int64)

func DownloadSABR(ctx context.Context, options SABROptions, outputPath string, progress ProgressFunc) error {
	manifestURL, err := sabrURL(options)
	if err != nil {
		return err
	}
	args := sabrInputArgs()
	if authorization := strings.TrimSpace(options.Authorization); authorization != "" {
		args = append(args, "-headers", "Authorization: "+authorization+"\r\n")
	}
	args = append(args, "-i", manifestURL)
	if options.AudioOnly {
		args = append(args, "-map", "0:a:0", "-vn")
	} else {
		args = append(args, "-map", "0:v:0", "-map", "0:a:0")
	}
	args = append(args, "-c", "copy", "-progress", "pipe:1", "-nostats", outputPath)
	command := exec.CommandContext(ctx, "ffmpeg", args...)
	stdout, err := command.StdoutPipe()
	if err != nil {
		return fmt.Errorf("open ffmpeg progress: %w", err)
	}
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Start(); err != nil {
		return fmt.Errorf("start ffmpeg SABR download: %w", err)
	}
	progressErr := readProgress(stdout, progress)
	waitErr := command.Wait()
	if waitErr != nil {
		return fmt.Errorf("ffmpeg SABR download failed: %w: %s", waitErr, strings.TrimSpace(stderr.String()))
	}
	if progressErr != nil {
		return fmt.Errorf("read ffmpeg SABR progress: %w", progressErr)
	}
	return nil
}

func readProgress(reader io.Reader, progress ProgressFunc) error {
	if progress == nil {
		_, err := io.Copy(io.Discard, reader)
		return err
	}
	scanner := bufio.NewScanner(reader)
	lastBytes := int64(-1)
	for scanner.Scan() {
		key, value, found := strings.Cut(scanner.Text(), "=")
		if !found || key != "total_size" {
			continue
		}
		downloadedBytes, err := strconv.ParseInt(value, 10, 64)
		if err != nil || downloadedBytes < 0 || downloadedBytes == lastBytes {
			continue
		}
		lastBytes = downloadedBytes
		progress(downloadedBytes)
	}
	return scanner.Err()
}

func sabrInputArgs() []string {
	return []string{
		"-y", "-hide_banner", "-loglevel", "error",
		"-reconnect", "1",
		"-reconnect_streamed", "1",
		"-reconnect_on_network_error", "1",
		"-reconnect_on_http_error", "4xx,5xx",
		"-reconnect_delay_max", "10",
	}
}

func sabrURL(options SABROptions) (string, error) {
	parsed, err := url.Parse(options.ManifestURL)
	if err != nil {
		return "", err
	}
	query := parsed.Query()
	query.Set("workload", "download")
	query.Set("audioItag", strconv.Itoa(options.AudioItag))
	if options.VideoItag > 0 {
		query.Set("videoItag", strconv.Itoa(options.VideoItag))
	}
	if options.AudioTrackID != "" {
		query.Set("audioTrackId", options.AudioTrackID)
	}
	if options.AudioOnly {
		query.Set("audioOnly", "true")
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}
