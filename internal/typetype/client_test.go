package typetype

import "testing"

func TestNormalizeWatchURLUnwrapsTypeTypeWatchURL(t *testing.T) {
	input := "https://watch.eltux.fr/watch?v=https%3A%2F%2Fwww.youtube.com%2Fwatch%3Fv%3DdQw4w9WgXcQ"
	got, err := NormalizeWatchURL(input)
	if err != nil {
		t.Fatal(err)
	}
	want := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	if got != want {
		t.Fatalf("NormalizeWatchURL() = %q, want %q", got, want)
	}
}

func TestNormalizeWatchURLKeepsDirectURL(t *testing.T) {
	input := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	got, err := NormalizeWatchURL(input)
	if err != nil {
		t.Fatal(err)
	}
	if got != input {
		t.Fatalf("NormalizeWatchURL() = %q, want %q", got, input)
	}
}
