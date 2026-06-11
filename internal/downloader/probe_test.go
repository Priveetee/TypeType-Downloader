package downloader

import "testing"

func TestParseContentRangeSize(t *testing.T) {
	size, err := parseContentRangeSize("bytes 0-0/12345")
	if err != nil {
		t.Fatal(err)
	}
	if size != 12345 {
		t.Fatalf("size = %d, want 12345", size)
	}
}

func TestParseContentRangeSizeRejectsUnknownTotal(t *testing.T) {
	if _, err := parseContentRangeSize("bytes 0-0/*"); err == nil {
		t.Fatal("expected error")
	}
}
