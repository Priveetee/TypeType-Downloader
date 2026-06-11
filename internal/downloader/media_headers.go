package downloader

import (
	"net/http"
	"net/url"
	"strings"
)

const desktopUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/122 Safari/537.36"

func effectiveRangeMode(rawURL string, configured string) string {
	if requiresHeaderRange(rawURL) {
		return "header"
	}
	return configured
}

func requiresHeaderRange(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	path := strings.TrimPrefix(parsed.EscapedPath(), "/")
	if path == "proxy" || path == "proxy/nicovideo" {
		return true
	}
	host := strings.ToLower(parsed.Hostname())
	return isBilibiliHost(host) || isNicoHost(host)
}

func applyMediaHeaders(req *http.Request, rawURL string) {
	for name, values := range MediaHeaders(rawURL, req.Header.Get("Range") != "") {
		for _, value := range values {
			req.Header.Set(name, value)
		}
	}
}

func MediaHeaders(rawURL string, rangeRequest bool) http.Header {
	headers := http.Header{}
	headers.Set("User-Agent", desktopUserAgent)
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return headers
	}
	host := strings.ToLower(parsed.Hostname())
	if isBilibiliHost(host) {
		headers.Set("Referer", "https://www.bilibili.com")
		headers.Set("Accept", "*/*")
		if rangeRequest {
			headers.Set("Connection", "close")
		}
	}
	if isNicoHost(host) {
		if bid := nicoDomandBid(parsed.Fragment); bid != "" {
			headers.Set("Cookie", "domand_bid="+bid)
		}
	}
	return headers
}

func isBilibiliHost(host string) bool {
	return strings.Contains(host, "bilibili") ||
		strings.Contains(host, "bilivideo") ||
		strings.HasSuffix(host, "hdslb.com") ||
		strings.Contains(host, "akamaized")
}

func isNicoHost(host string) bool {
	return strings.HasSuffix(host, "nicovideo.jp")
}

func nicoDomandBid(fragment string) string {
	decoded, err := url.QueryUnescape(fragment)
	if err != nil {
		decoded = fragment
	}
	for _, item := range strings.Split(decoded, "&") {
		cookie := strings.TrimPrefix(item, "cookie=")
		if cookie == item || !strings.HasPrefix(cookie, "domand_bid=") {
			continue
		}
		return strings.TrimPrefix(cookie, "domand_bid=")
	}
	return ""
}
