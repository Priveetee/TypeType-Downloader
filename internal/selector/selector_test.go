package selector

import (
	"testing"

	"typetype-downloader-go/internal/typetype"
)

func TestSelectMP4WithOptionsChoosesHighestAllowedHeightAndPreferredAudio(t *testing.T) {
	avc := "avc1.640028"
	mp4a := "mp4a.40.2"
	bitrate := 128000
	trackID := "orig"
	stream := &typetype.StreamResponse{
		Title:                   "title",
		PreferredDefaultAudioID: &trackID,
		VideoOnlyStreams: []typetype.VideoStreamItem{
			{URL: "v720", MimeType: "video/mp4", Codec: &avc, Itag: 136, Height: 720, ContentLength: 100},
			{URL: "v1080", MimeType: "video/mp4", Codec: &avc, Itag: 137, Height: 1080, ContentLength: 200},
			{URL: "v1440", MimeType: "video/mp4", Codec: &avc, Itag: 271, Height: 1440, ContentLength: 300},
		},
		AudioStreams: []typetype.AudioStreamItem{
			{URL: "a1", MimeType: "audio/mp4", Codec: &mp4a, Itag: 140, Bitrate: &bitrate, ContentLength: 50},
			{URL: "a2", MimeType: "audio/mp4", Codec: &mp4a, Itag: 141, AudioTrackID: &trackID, ContentLength: 40},
		},
	}
	selection, err := SelectMP4WithOptions(stream, Options{Container: "mp4", MaxHeight: 1080})
	if err != nil {
		t.Fatal(err)
	}
	if selection.Video.Itag != 137 {
		t.Fatalf("video itag = %d, want 137", selection.Video.Itag)
	}
	if selection.Audio.Itag != 141 {
		t.Fatalf("audio itag = %d, want 141", selection.Audio.Itag)
	}
}

func TestSelectMP4WithOptionsHonorsExplicitItags(t *testing.T) {
	avc := "avc1.640028"
	mp4a := "mp4a.40.2"
	stream := &typetype.StreamResponse{
		VideoOnlyStreams: []typetype.VideoStreamItem{
			{URL: "v1", MimeType: "video/mp4", Codec: &avc, Itag: 136, Height: 720, ContentLength: 100},
			{URL: "v2", MimeType: "video/mp4", Codec: &avc, Itag: 137, Height: 1080, ContentLength: 200},
		},
		AudioStreams: []typetype.AudioStreamItem{
			{URL: "a1", MimeType: "audio/mp4", Codec: &mp4a, Itag: 140, ContentLength: 50},
			{URL: "a2", MimeType: "audio/mp4", Codec: &mp4a, Itag: 141, ContentLength: 40},
		},
	}
	selection, err := SelectMP4WithOptions(stream, Options{Container: "mp4", VideoItag: 136, AudioItag: 140})
	if err != nil {
		t.Fatal(err)
	}
	if selection.Video.Itag != 136 || selection.Audio.Itag != 140 {
		t.Fatalf("selection = %d+%d, want 136+140", selection.Video.Itag, selection.Audio.Itag)
	}
}

func TestSelectMP4WithOptionsAcceptsSABRManifest(t *testing.T) {
	avc := "avc1.640028"
	mp4a := "mp4a.40.2"
	stream := &typetype.StreamResponse{
		VideoOnlyStreams: []typetype.VideoStreamItem{{
			MimeType: "video/mp4", Codec: &avc, Itag: 137, Height: 1080,
			DeliveryMethod: "sabr", ManifestURL: "/sabr/manifest/video-id",
		}},
		AudioStreams: []typetype.AudioStreamItem{{
			MimeType: "audio/mp4", Codec: &mp4a, Itag: 140,
			DeliveryMethod: "sabr", ManifestURL: "/sabr/manifest/video-id",
		}},
	}
	selection, err := SelectMP4WithOptions(stream, Options{Container: "mp4", MaxHeight: 1080})
	if err != nil {
		t.Fatal(err)
	}
	if selection.Video.Itag != 137 || selection.Audio.Itag != 140 {
		t.Fatalf("selection = %d+%d, want 137+140", selection.Video.Itag, selection.Audio.Itag)
	}
}

func TestSelectMP4WithOptionsRemuxesSABRVP9AndAACToMP4(t *testing.T) {
	vp9 := "vp9"
	mp4a := "mp4a.40.2"
	stream := &typetype.StreamResponse{
		VideoOnlyStreams: []typetype.VideoStreamItem{{
			MimeType: "video/webm", Codec: &vp9, Itag: 308, Height: 1440,
			DeliveryMethod: "sabr", ManifestURL: "/sabr/manifest/video-id",
		}},
		AudioStreams: []typetype.AudioStreamItem{{
			MimeType: "audio/mp4", Codec: &mp4a, Itag: 140,
			DeliveryMethod: "sabr", ManifestURL: "/sabr/manifest/video-id",
		}},
	}
	selection, err := SelectMP4WithOptions(stream, Options{Container: "webm", MaxHeight: 1440})
	if err != nil {
		t.Fatal(err)
	}
	if selection.Video.Itag != 308 || selection.Audio.Itag != 140 || selection.Container != "mp4" {
		t.Fatalf("selection = %d+%d/%s, want 308+140/mp4", selection.Video.Itag, selection.Audio.Itag, selection.Container)
	}
}

func TestSelectMP4WithOptionsAllowsUnknownSizeAndCodec(t *testing.T) {
	mp4a := "mp4a"
	stream := &typetype.StreamResponse{
		VideoOnlyStreams: []typetype.VideoStreamItem{
			{URL: "video.m3u8", MimeType: "video/mp4", Height: 360},
		},
		AudioStreams: []typetype.AudioStreamItem{
			{URL: "audio.m3u8", MimeType: "audio/mp4", Codec: &mp4a},
		},
	}
	selection, err := SelectMP4WithOptions(stream, Options{Container: "mp4", MaxHeight: 360})
	if err != nil {
		t.Fatal(err)
	}
	if selection.Video.URL != "video.m3u8" || selection.Audio.URL != "audio.m3u8" {
		t.Fatalf("selection = %s+%s", selection.Video.URL, selection.Audio.URL)
	}
}

func TestSelectAudioOnlyDefaultsToM4A(t *testing.T) {
	mp4a := "mp4a.40.2"
	opus := "opus"
	stream := &typetype.StreamResponse{
		Title: "audio title",
		AudioStreams: []typetype.AudioStreamItem{
			{URL: "webm", MimeType: "audio/webm", Codec: &opus, Itag: 251, ContentLength: 90},
			{URL: "m4a", MimeType: "audio/mp4", Codec: &mp4a, Itag: 140, ContentLength: 80},
		},
	}
	selection, err := SelectAudioOnly(stream, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if selection.Container != "m4a" || selection.Audio.Itag != 140 {
		t.Fatalf("selection = %s/%d, want m4a/140", selection.Container, selection.Audio.Itag)
	}
}

func TestSelectAudioOnlyCanSelectWebMOpus(t *testing.T) {
	mp4a := "mp4a.40.2"
	opus := "opus"
	stream := &typetype.StreamResponse{
		AudioStreams: []typetype.AudioStreamItem{
			{URL: "m4a", MimeType: "audio/mp4", Codec: &mp4a, Itag: 140, ContentLength: 80},
			{URL: "webm", MimeType: "audio/webm", Codec: &opus, Itag: 251, ContentLength: 90},
		},
	}
	selection, err := SelectAudioOnly(stream, Options{Container: "webm"})
	if err != nil {
		t.Fatal(err)
	}
	if selection.Container != "webm" || selection.Audio.Itag != 251 {
		t.Fatalf("selection = %s/%d, want webm/251", selection.Container, selection.Audio.Itag)
	}
}

func TestSelectMP4WithOptionsAcceptsUnknownRemoteMetadata(t *testing.T) {
	stream := &typetype.StreamResponse{
		VideoOnlyStreams: []typetype.VideoStreamItem{{URL: "video.m3u8", MimeType: "video/mp4", Itag: -1}},
		AudioStreams:     []typetype.AudioStreamItem{{URL: "audio.m3u8", MimeType: "audio/mp4", Itag: -1}},
	}
	selection, err := SelectMP4WithOptions(stream, Options{Container: "mp4"})
	if err != nil {
		t.Fatal(err)
	}
	if selection.Video.URL != "video.m3u8" || selection.Audio.URL != "audio.m3u8" {
		t.Fatalf("selection = %s + %s", selection.Video.URL, selection.Audio.URL)
	}
}
