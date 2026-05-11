package artifact

import (
	"context"
	"net/http"
	"os"
)

type LocalStore struct{}

func NewLocalStore() *LocalStore { return &LocalStore{} }

func (s *LocalStore) Name() string { return "artifact" }

func (s *LocalStore) Health(_ context.Context) error { return nil }

func (s *LocalStore) Save(_ context.Context, localPath string, _ string) (Saved, error) {
	return Saved{Backend: "local", Location: localPath}, nil
}

func (s *LocalStore) ServeHTTP(w http.ResponseWriter, r *http.Request, saved Saved, _ string) error {
	if _, err := os.Stat(saved.Location); err != nil {
		return err
	}
	http.ServeFile(w, r, saved.Location)
	return nil
}

func (s *LocalStore) Delete(_ context.Context, saved Saved) error {
	if saved.Location == "" {
		return nil
	}
	if err := os.Remove(saved.Location); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
