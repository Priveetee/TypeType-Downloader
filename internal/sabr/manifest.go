package sabr

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"strings"
)

type manifest struct {
	Period period `xml:"Period"`
}

type period struct {
	AdaptationSets []adaptationSet `xml:"AdaptationSet"`
}

type adaptationSet struct {
	MimeType        string           `xml:"mimeType,attr"`
	Representations []representation `xml:"Representation"`
}

type representation struct {
	Segments segmentList `xml:"SegmentList"`
}

type segmentList struct {
	Initialization initialization `xml:"Initialization"`
	Segments       []segment      `xml:"SegmentURL"`
}

type initialization struct {
	SourceURL string `xml:"sourceURL,attr"`
}

type segment struct {
	Media string `xml:"media,attr"`
}

type Track struct {
	Kind string
	URLs []string
}

func parseManifest(reader io.Reader, base *url.URL, audioOnly bool) ([]Track, error) {
	var document manifest
	if err := xml.NewDecoder(reader).Decode(&document); err != nil {
		return nil, fmt.Errorf("decode SABR manifest: %w", err)
	}
	tracks := make([]Track, 0, 2)
	for _, adaptation := range document.Period.AdaptationSets {
		kind := strings.TrimSuffix(adaptation.MimeType, "/mp4")
		if kind != "audio" && (kind != "video" || audioOnly) {
			continue
		}
		if len(adaptation.Representations) == 0 {
			return nil, fmt.Errorf("SABR manifest %s track has no representation", kind)
		}
		list := adaptation.Representations[0].Segments
		refs := make([]string, 0, len(list.Segments)+1)
		refs = append(refs, list.Initialization.SourceURL)
		for _, item := range list.Segments {
			refs = append(refs, item.Media)
		}
		urls, err := resolveURLs(base, refs)
		if err != nil {
			return nil, fmt.Errorf("resolve SABR %s track: %w", kind, err)
		}
		tracks = append(tracks, Track{Kind: kind, URLs: urls})
	}
	want := 2
	if audioOnly {
		want = 1
	}
	if len(tracks) != want {
		return nil, fmt.Errorf("SABR manifest has %d usable tracks, want %d", len(tracks), want)
	}
	return tracks, nil
}

func resolveURLs(base *url.URL, refs []string) ([]string, error) {
	urls := make([]string, 0, len(refs))
	for _, raw := range refs {
		if strings.TrimSpace(raw) == "" {
			return nil, fmt.Errorf("empty segment URL")
		}
		ref, err := url.Parse(raw)
		if err != nil {
			return nil, err
		}
		urls = append(urls, base.ResolveReference(ref).String())
	}
	return urls, nil
}
