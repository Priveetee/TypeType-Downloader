package pipeline

import (
	"context"
	"fmt"
	"testing"

	"typetype-downloader-go/internal/storage"
)

func TestFailureCodePreservesStorageFailure(t *testing.T) {
	err := fmt.Errorf("reserve: %w", storage.ErrInsufficientStorage)
	if got := failureCode(context.Background(), err); got != "insufficient_storage" {
		t.Fatalf("failure code = %q", got)
	}
}
