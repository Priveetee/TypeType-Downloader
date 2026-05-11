package job

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

type cacheInput struct {
	URL     string  `json:"url"`
	Options Options `json:"options"`
}

func CacheKey(rawURL string, options Options) (string, error) {
	payload, err := json.Marshal(cacheInput{URL: strings.TrimSpace(rawURL), Options: normalizeOptions(options)})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}

func normalizeOptions(options Options) Options {
	options.Mode = strings.TrimSpace(options.Mode)
	options.Quality = strings.TrimSpace(options.Quality)
	options.Format = strings.TrimSpace(options.Format)
	options.Container = strings.TrimSpace(strings.ToLower(options.Container))
	options.VideoCodec = strings.TrimSpace(strings.ToLower(options.VideoCodec))
	options.AudioCodec = strings.TrimSpace(strings.ToLower(options.AudioCodec))
	options.VideoItag = strings.TrimSpace(options.VideoItag)
	options.AudioItag = strings.TrimSpace(options.AudioItag)
	for i := range options.SponsorBlockCategories {
		options.SponsorBlockCategories[i] = strings.TrimSpace(options.SponsorBlockCategories[i])
	}
	return options
}
