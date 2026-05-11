package artifact

import (
	"context"
	"net/http"
	"time"
)

type Saved struct {
	Backend  string
	Location string
	Expires  time.Time
}

type Store interface {
	Name() string
	Health(ctx context.Context) error
	Save(ctx context.Context, localPath string, objectKey string) (Saved, error)
	ServeHTTP(w http.ResponseWriter, r *http.Request, saved Saved, fileName string) error
	Delete(ctx context.Context, saved Saved) error
}
