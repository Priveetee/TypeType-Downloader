package ffmpeg

import (
	"context"
	"fmt"
	"os/exec"
)

func MergeCopy(ctx context.Context, videoPath string, audioPath string, outputPath string) error {
	cmd := exec.CommandContext(ctx, "ffmpeg", "-y", "-hide_banner", "-loglevel", "error", "-i", videoPath, "-i", audioPath, "-c", "copy", "-map", "0:v:0", "-map", "1:a:0", outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg merge failed: %w: %s", err, string(output))
	}
	return nil
}
