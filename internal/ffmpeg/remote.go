package ffmpeg

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

func DownloadRemote(ctx context.Context, videoURL string, audioURL string, outputPath string) error {
	args := []string{"-y", "-hide_banner", "-loglevel", "error"}
	args = append(args, remoteInput(videoURL)...)
	args = append(args, remoteInput(audioURL)...)
	args = append(args, "-c", "copy", "-map", "0:v:0", "-map", "1:a:0", outputPath)
	output, err := exec.CommandContext(ctx, "ffmpeg", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg remote download failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func DownloadRemoteAudio(ctx context.Context, audioURL string, outputPath string) error {
	args := remoteAudioArgs(audioURL, outputPath)
	output, err := exec.CommandContext(ctx, "ffmpeg", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg remote audio download failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func remoteAudioArgs(audioURL string, outputPath string) []string {
	args := []string{"-y", "-hide_banner", "-loglevel", "error"}
	args = append(args, remoteInput(audioURL)...)
	return append(args, "-c", "copy", "-map", "0:a:0", "-vn", outputPath)
}

func remoteInput(rawURL string) []string {
	args := []string{"-user_agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/122 Safari/537.36"}
	if strings.Contains(rawURL, "/proxy/nicovideo") {
		args = append(args, "-extension_picky", "0", "-allowed_segment_extensions", "ALL")
	}
	parsed, err := url.Parse(rawURL)
	if err == nil && (strings.Contains(parsed.Host, "bilivideo") || strings.Contains(parsed.Host, "akamaized")) {
		args = append(args, "-headers", "Referer: https://www.bilibili.com\r\n")
	}
	return append(args, "-i", rawURL)
}
