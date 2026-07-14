package sabr

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func Download(ctx context.Context, client *http.Client, options Options, progress ProgressFunc) error {
	manifestURL, err := buildManifestURL(options)
	if err != nil {
		return err
	}
	tracks, err := fetchManifest(ctx, client, manifestURL, options.Authorization, options.AudioOnly)
	if err != nil {
		return err
	}
	tempDir, err := os.MkdirTemp(options.WorkDir, "sabr-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)
	files, plans := planFiles(tracks, tempDir, options)
	reporter := newReporter(progress)
	if err := downloadFiles(ctx, client, files, options.Authorization, options.Workers, reporter); err != nil {
		return err
	}
	for _, plan := range plans {
		if err := assemble(ctx, plan.Target, plan.Parts); err != nil {
			return err
		}
	}
	reporter.finish()
	return nil
}

func fetchManifest(ctx context.Context, client *http.Client, rawURL string, authorization string, audioOnly bool) ([]Track, error) {
	var last error
	for attempt := 1; attempt <= 4; attempt++ {
		response, err := request(ctx, client, rawURL, authorization)
		if err == nil {
			tracks, parseErr := parseManifest(response.Body, response.Request.URL, audioOnly)
			response.Body.Close()
			if parseErr != nil {
				return nil, parseErr
			}
			return tracks, nil
		}
		last = err
		if attempt < 4 {
			if err := retryDelay(ctx, attempt); err != nil {
				return nil, err
			}
		}
	}
	return nil, fmt.Errorf("fetch SABR manifest failed after 4 attempts: %w", last)
}

func buildManifestURL(options Options) (string, error) {
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

func request(ctx context.Context, client *http.Client, rawURL string, authorization string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	if value := strings.TrimSpace(authorization); value != "" {
		req.Header.Set("Authorization", value)
	}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		response.Body.Close()
		return nil, fmt.Errorf("GET %s returned %s", req.URL.Path, response.Status)
	}
	return response, nil
}

func planFiles(tracks []Track, tempDir string, options Options) ([]filePlan, []trackPlan) {
	plans := make([]trackPlan, 0, len(tracks))
	for trackIndex, track := range tracks {
		target := options.AudioPath
		if track.Kind == "video" {
			target = options.VideoPath
		}
		parts := make([]string, 0, len(track.URLs))
		for partIndex := range track.URLs {
			path := filepath.Join(tempDir, fmt.Sprintf("%d-%06d.part", trackIndex, partIndex))
			parts = append(parts, path)
		}
		plans = append(plans, trackPlan{Parts: parts, Target: target})
	}
	files := make([]filePlan, 0)
	for partIndex := 0; ; partIndex++ {
		added := false
		for trackIndex, track := range tracks {
			if partIndex >= len(track.URLs) {
				continue
			}
			files = append(files, filePlan{URL: track.URLs[partIndex], Path: plans[trackIndex].Parts[partIndex]})
			added = true
		}
		if !added {
			break
		}
	}
	return files, plans
}
