package typetype

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (c *Client) FetchStream(ctx context.Context, rawURL string, authorization string) (*StreamResponse, error) {
	resolved, err := NormalizeWatchURL(rawURL)
	if err != nil {
		return nil, err
	}
	endpoint := c.baseURL + "/streams?url=" + url.QueryEscape(resolved)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	if authorization = strings.TrimSpace(authorization); authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return nil, fmt.Errorf("streams endpoint returned %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}
	var stream StreamResponse
	if err := json.NewDecoder(res.Body).Decode(&stream); err != nil {
		return nil, err
	}
	return &stream, nil
}

func (c *Client) MediaURL(path string) string {
	trimmed := strings.TrimSpace(path)
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return trimmed
	}
	return c.baseURL + "/" + strings.TrimLeft(trimmed, "/")
}

func (c *Client) ProxyMediaURL(rawURL string) string {
	path := "/proxy"
	if parsed, err := url.Parse(rawURL); err == nil && strings.Contains(parsed.Host, "nicovideo.jp") {
		path = "/proxy/nicovideo"
	}
	return c.baseURL + path + "?url=" + url.QueryEscape(rawURL)
}

func NormalizeWatchURL(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid URL: %s", rawURL)
	}
	wrapped := parsed.Query().Get("v")
	if parsed.Path == "/watch" && wrapped != "" {
		inner, err := url.QueryUnescape(wrapped)
		if err != nil {
			return "", err
		}
		if innerURL, err := url.Parse(inner); err == nil && innerURL.Scheme != "" && innerURL.Host != "" {
			return inner, nil
		}
	}
	return rawURL, nil
}
