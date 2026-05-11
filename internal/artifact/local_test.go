package artifact

import (
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalStoreSaveServeDelete(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.mp4")
	if err := os.WriteFile(path, []byte("media"), 0o644); err != nil {
		t.Fatal(err)
	}
	store := NewLocalStore()
	saved, err := store.Save(context.Background(), path, "ignored")
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/artifact", nil)
	if err := store.ServeHTTP(recorder, request, saved, "out.mp4"); err != nil {
		t.Fatal(err)
	}
	if recorder.Body.String() != "media" {
		t.Fatalf("body = %q", recorder.Body.String())
	}
	if err := store.Delete(context.Background(), saved); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected deleted file, got %v", err)
	}
}
