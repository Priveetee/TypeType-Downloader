package ffmpeg

import (
	"context"
	"fmt"
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

func DownloadSABR(ctx context.Context, options SABROptions, outputPath string) error {
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
	args = append(args, "-c", "copy", outputPath)
	output, err := exec.CommandContext(ctx, "ffmpeg", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg SABR download failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
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
