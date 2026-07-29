package downloader

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

type chunk struct {
	start int64
	end   int64
}

func downloadChunk(ctx context.Context, client *http.Client, file *os.File, source Source, part chunk, options Options, downloaded *atomic.Int64, started time.Time, progress ProgressFunc) error {
	var lastErr error
	for attempt := 0; attempt < options.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt) * 300 * time.Millisecond):
			}
		}
		err := fetchChunk(ctx, client, file, source, part, options, downloaded, started, progress)
		if err == nil {
			return nil
		}
		lastErr = err
	}
	return fmt.Errorf("%s bytes %d-%d failed: %w", source.Name, part.start, part.end, lastErr)
}
