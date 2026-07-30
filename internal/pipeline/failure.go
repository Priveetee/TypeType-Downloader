package pipeline

import (
	"context"
	"errors"

	"typetype-downloader-go/internal/storage"
)

func failureCode(ctx context.Context, err error) string {
	if ctx.Err() != nil || errors.Is(err, context.Canceled) {
		return "cancelled"
	}
	if errors.Is(err, storage.ErrInsufficientStorage) {
		return "insufficient_storage"
	}
	return "download_failed"
}
