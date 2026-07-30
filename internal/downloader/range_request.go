package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func fetchChunk(
	ctx context.Context,
	client *http.Client,
	file *os.File,
	source Source,
	part chunk,
	configuredMode string,
	buffer []byte,
) error {
	rangeMode := effectiveRangeMode(source.URL, configuredMode)
	request, err := rangedRequest(ctx, source.URL, part, rangeMode)
	if err != nil {
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if err := validateRangeResponse(response, part, rangeMode); err != nil {
		return err
	}
	if err := copyChunk(response.Body, file, part, buffer); err != nil {
		return err
	}
	if response.ContentLength < 0 {
		var trailing [1]byte
		if n, readErr := response.Body.Read(trailing[:]); n > 0 || readErr != io.EOF {
			return fmt.Errorf("chunk overflow")
		}
	}
	return nil
}

func rangedRequest(ctx context.Context, rawURL string, part chunk, rangeMode string) (*http.Request, error) {
	requestURL, err := rangedURL(rawURL, part, rangeMode)
	if err != nil {
		return nil, err
	}
	requestURL, err = stripFragment(requestURL)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept-Encoding", "identity")
	if rangeMode == "header" {
		request.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", part.start, part.end))
	}
	applyMediaHeaders(request, rawURL)
	return request, nil
}

func validateRangeResponse(response *http.Response, part chunk, configuredMode string) error {
	expected := part.end - part.start + 1
	if response.StatusCode != http.StatusPartialContent &&
		!(configuredMode == "query" && response.StatusCode == http.StatusOK) {
		return fmt.Errorf("unexpected HTTP status %d", response.StatusCode)
	}
	if configuredMode == "header" {
		want := fmt.Sprintf("bytes %d-%d/", part.start, part.end)
		if contentRange := response.Header.Get("Content-Range"); !strings.HasPrefix(contentRange, want) {
			return fmt.Errorf("unexpected Content-Range %q", contentRange)
		}
	}
	if response.ContentLength >= 0 && response.ContentLength != expected {
		return fmt.Errorf("unexpected content length %d, expected %d", response.ContentLength, expected)
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
