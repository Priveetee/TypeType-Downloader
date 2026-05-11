package pipeline

import (
	"context"
	"fmt"
	"time"
)

func retry(ctx context.Context, attempts int, label string, run func() error) error {
	if attempts <= 0 {
		attempts = 1
	}
	var last error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := run(); err != nil {
			last = err
		} else {
			return nil
		}
		if attempt < attempts {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt) * 500 * time.Millisecond):
			}
		}
	}
	return fmt.Errorf("%s failed after %d attempts: %w", label, attempts, last)
}
