package sabr

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func openPartTracks(options Options, part int) ([]streamTrack, error) {
	specs := trackSpecs(options)
	tracks := make([]streamTrack, 0, len(specs))
	for _, spec := range specs {
		if spec.itag <= 0 || strings.TrimSpace(spec.path) == "" {
			closeTracks(tracks)
			return nil, fmt.Errorf("missing SABR %s selection or target path", spec.kind)
		}
		spec.path = partPath(spec.path, part)
		output, err := os.Create(spec.path)
		if err != nil {
			closeTracks(tracks)
			return nil, fmt.Errorf("create SABR %s part: %w", spec.kind, err)
		}
		spec.output = output
		tracks = append(tracks, spec)
	}
	return tracks, nil
}

func assembleDownload(options Options, parts int) error {
	specs := trackSpecs(options)
	for index := range specs {
		if err := assembleTrack(&specs[index], parts); err != nil {
			return err
		}
	}
	for index := range specs {
		if err := os.Rename(specs[index].path+".download", specs[index].path); err != nil {
			return fmt.Errorf("commit SABR %s track: %w", specs[index].kind, err)
		}
	}
	return nil
}

func assembleTrack(track *streamTrack, parts int) error {
	output, err := os.Create(track.path + ".download")
	if err != nil {
		return fmt.Errorf("create SABR %s track: %w", track.kind, err)
	}
	ok := false
	defer func() {
		_ = output.Close()
		if !ok {
			_ = os.Remove(track.path + ".download")
		}
	}()
	for part := range parts {
		input, err := os.Open(partPath(track.path, part))
		if err != nil {
			return fmt.Errorf("open SABR %s part: %w", track.kind, err)
		}
		_, copyErr := io.Copy(output, input)
		closeErr := input.Close()
		if copyErr != nil {
			return fmt.Errorf("assemble SABR %s track: %w", track.kind, copyErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close SABR %s part: %w", track.kind, closeErr)
		}
	}
	if err := output.Close(); err != nil {
		return fmt.Errorf("close SABR %s track: %w", track.kind, err)
	}
	ok = true
	return nil
}

func cleanupDownloadFiles(options Options, parts int) {
	for _, track := range trackSpecs(options) {
		_ = os.Remove(track.path + ".download")
		for part := range parts {
			_ = os.Remove(partPath(track.path, part))
		}
	}
}

func closeTracks(tracks []streamTrack) error {
	var first error
	for index := range tracks {
		if tracks[index].output == nil {
			continue
		}
		if err := tracks[index].output.Close(); err != nil && first == nil {
			first = fmt.Errorf("close SABR %s part: %w", tracks[index].kind, err)
		}
		tracks[index].output = nil
	}
	return first
}

func trackSpecs(options Options) []streamTrack {
	specs := []streamTrack{{kind: "audio", itag: options.AudioItag, path: options.AudioPath}}
	if !options.AudioOnly {
		specs = append(specs, streamTrack{kind: "video", itag: options.VideoItag, path: options.VideoPath})
	}
	return specs
}

func partPath(target string, part int) string {
	return fmt.Sprintf("%s.download.part-%02d", target, part)
}
