package downloader

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func probeSourceSize(ctx context.Context, client *http.Client, rawURL string, mode string) (int64, error) {
	rangeMode := effectiveRangeMode(rawURL, mode)
	requestURL, err := rangedURL(rawURL, chunk{start: 0, end: 0}, rangeMode)
	if err != nil {
		return 0, err
	}
	requestURL, err = stripFragment(requestURL)
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept-Encoding", "identity")
	if rangeMode == "header" {
		req.Header.Set("Range", "bytes=0-0")
	}
	applyMediaHeaders(req, rawURL)
	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	if rangeMode == "header" && res.StatusCode == http.StatusPartialContent {
		return parseContentRangeSize(res.Header.Get("Content-Range"))
	}
	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected HTTP status %d", res.StatusCode)
	}
	if size := res.ContentLength; size > 0 {
		return size, nil
	}
	return 0, fmt.Errorf("missing content length")
}

func parseContentRangeSize(value string) (int64, error) {
	_, total, ok := strings.Cut(value, "/")
	if !ok || total == "" || total == "*" {
		return 0, fmt.Errorf("unexpected Content-Range %q", value)
	}
	size, err := strconv.ParseInt(total, 10, 64)
	if err != nil || size <= 0 {
		return 0, fmt.Errorf("unexpected Content-Range %q", value)
	}
	return size, nil
}
