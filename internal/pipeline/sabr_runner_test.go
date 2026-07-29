package pipeline

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"

	"typetype-downloader-go/internal/artifact"
	"typetype-downloader-go/internal/job"
)

func TestRunSABRArtifactRecordsProcessingTimings(t *testing.T) {
	store := job.NewStore("http://localhost")
	store.Restore([]*job.Record{{ID: "job", Status: job.StatusRunning}})
	runner := &Runner{store: store, storage: timingArtifactStore{}}
	root := t.TempDir()
	paths := artifact.Paths{
		WorkDir: filepath.Join(root, "work"),
		Output:  filepath.Join(root, "output.mp4"),
		Key:     "artifact",
	}

	err := runner.runSABRArtifact(t.Context(), "job", paths, func() (artifactTimings, error) {
		return artifactTimings{downloadMs: 123, muxMs: 45}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	record, ok := store.Get("job")
	if !ok || record.DownloadMs == nil || *record.DownloadMs != 123 ||
		record.MuxMs == nil || *record.MuxMs != 45 {
		t.Fatalf("record = %#v", record)
	}
}

type timingArtifactStore struct{}

func (timingArtifactStore) Name() string {
	return "test"
}

func (timingArtifactStore) Health(context.Context) error {
	return nil
}

func (timingArtifactStore) Save(context.Context, string, string) (artifact.Saved, error) {
	return artifact.Saved{Backend: "test", Location: "artifact"}, nil
}

func (timingArtifactStore) ServeHTTP(http.ResponseWriter, *http.Request, artifact.Saved, string) error {
	return nil
}

func (timingArtifactStore) Delete(context.Context, artifact.Saved) error {
	return nil
}
