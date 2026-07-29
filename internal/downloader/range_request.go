package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

func fetchChunk(ctx context.Context, client *http.Client, file *os.File, source Source, part chunk, options Options, downloaded *atomic.Int64, started time.Time, progress ProgressFunc) error {
	rangeMode := effectiveRangeMode(source.URL, options.RangeMode)
	requestURL, err := rangedURL(source.URL, part, rangeMode)
	if err != nil {
		return err
	}
	requestURL, err = stripFragment(requestURL)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept-Encoding", "identity")
	if rangeMode == "header" {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", part.start, part.end))
	}
	applyMediaHeaders(req, source.URL)

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusPartialContent && !(rangeMode == "query" && res.StatusCode == http.StatusOK) {
		return fmt.Errorf("unexpected HTTP status %d", res.StatusCode)
	}
	if contentRange := res.Header.Get("Content-Range"); rangeMode == "header" && !strings.HasPrefix(contentRange, fmt.Sprintf("bytes %d-%d/", part.start, part.end)) {
		return fmt.Errorf("unexpected Content-Range %q", contentRange)
	}

	buf := make([]byte, options.BufferSize)
	position := part.start
	written := int64(0)
	for {
		n, readErr := res.Body.Read(buf)
		if n > 0 {
			if position+int64(n)-1 > part.end {
				return fmt.Errorf("chunk overflow")
			}
			if _, err := file.WriteAt(buf[:n], position); err != nil {
				return err
			}
			position += int64(n)
			written += int64(n)
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	if position != part.end+1 {
		return fmt.Errorf("short chunk: got %d expected %d", position-part.start, part.end-part.start+1)
	}
	current := downloaded.Add(written)
	if progress != nil {
		elapsed := time.Since(started).Seconds()
		if elapsed > 0 {
			progress(Progress{Name: source.Name, Downloaded: current, Total: source.Size, Speed: float64(current) / elapsed})
		}
	}
	return nil
}

func stripFragment(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	parsed.Fragment = ""
	return parsed.String(), nil
}

func rangedURL(rawURL string, part chunk, mode string) (string, error) {
	if mode == "header" {
		return rawURL, nil
	}
	if mode != "query" {
		return "", fmt.Errorf("unsupported range mode %q", mode)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	query := parsed.Query()
	query.Set("range", fmt.Sprintf("%d-%d", part.start, part.end))
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}
