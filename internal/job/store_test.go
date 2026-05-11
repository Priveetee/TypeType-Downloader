package job

import (
	"errors"
	"testing"
)

func TestStoreCreateStartProgressDone(t *testing.T) {
	store := NewStore("http://localhost")
	record, _, _, err := store.Create("https://example.com/watch?v=1", Options{})
	if err != nil {
		t.Fatal(err)
	}
	store.Start(record.ID, func() {})
	store.Progress(record.ID, Progress{Stage: "download", DownloadedBytes: 50, TotalBytes: 100, SpeedBytesPerSecond: 25})
	store.Done(record.ID, "/tmp/out.mp4", "local", nil, 10, 20)
	response, ok := store.Response(record.ID)
	if !ok {
		t.Fatal("missing response")
	}
	if response.Status != StatusDone {
		t.Fatalf("status = %s, want done", response.Status)
	}
	if response.ArtifactURL == nil || *response.ArtifactURL != "http://localhost/jobs/"+record.ID+"/artifact" {
		t.Fatalf("artifact URL = %#v", response.ArtifactURL)
	}
	if response.ProgressPercent == nil || *response.ProgressPercent != 100 {
		t.Fatalf("progress = %#v, want 100", response.ProgressPercent)
	}
}

func TestStoreCancelQueuedMarksFailed(t *testing.T) {
	store := NewStore("http://localhost")
	record, _, _, err := store.Create("https://example.com/watch?v=1", Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !store.Cancel(record.ID) {
		t.Fatal("Cancel() returned false")
	}
	response, _ := store.Response(record.ID)
	if response.Status != StatusFailed || response.ErrorCode == nil || *response.ErrorCode != "cancelled" {
		t.Fatalf("response = %#v", response)
	}
}

func TestStoreFailPublishesError(t *testing.T) {
	store := NewStore("http://localhost")
	record, _, _, err := store.Create("https://example.com/watch?v=1", Options{})
	if err != nil {
		t.Fatal(err)
	}
	store.Fail(record.ID, "boom", errors.New("failed"))
	response, _ := store.Response(record.ID)
	if response.Status != StatusFailed || response.Error == nil || *response.Error != "failed" {
		t.Fatalf("response = %#v", response)
	}
}

func TestStoreDeduplicatesDoneJobs(t *testing.T) {
	store := NewStore("http://localhost")
	first, cached, created, err := store.Create("https://example.com/watch?v=1", Options{Container: "mp4"})
	if err != nil {
		t.Fatal(err)
	}
	if cached || !created {
		t.Fatalf("first create cached=%v created=%v", cached, created)
	}
	store.Done(first.ID, "/tmp/out.mp4", "local", nil, 10, 20)
	second, cached, created, err := store.Create("https://example.com/watch?v=1", Options{Container: "mp4"})
	if err != nil {
		t.Fatal(err)
	}
	if second.ID != first.ID || !cached || created {
		t.Fatalf("second id=%s cached=%v created=%v", second.ID, cached, created)
	}
}

func TestStoreDeduplicatesRunningJobsWithoutCachedFlag(t *testing.T) {
	store := NewStore("http://localhost")
	first, _, _, err := store.Create("https://example.com/watch?v=1", Options{Container: "mp4"})
	if err != nil {
		t.Fatal(err)
	}
	store.Start(first.ID, func() {})
	second, cached, created, err := store.Create("https://example.com/watch?v=1", Options{Container: "mp4"})
	if err != nil {
		t.Fatal(err)
	}
	if second.ID != first.ID || cached || created {
		t.Fatalf("second id=%s cached=%v created=%v", second.ID, cached, created)
	}
}
