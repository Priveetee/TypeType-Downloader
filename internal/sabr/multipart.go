package sabr

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
)

func downloadParts(
	ctx context.Context,
	client *http.Client,
	options Options,
	parts int,
	progress *reporter,
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	failures := make(chan error, parts)
	var workers sync.WaitGroup
	for part := range parts {
		workers.Add(1)
		go func() {
			defer workers.Done()
			if err := downloadPart(ctx, client, options, part, parts, progress); err != nil {
				select {
				case failures <- err:
				default:
				}
				cancel()
			}
		}()
	}
	workers.Wait()
	close(failures)
	for err := range failures {
		if !errors.Is(err, context.Canceled) {
			return err
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

func downloadPart(
	ctx context.Context,
	client *http.Client,
	options Options,
	part int,
	parts int,
	progress *reporter,
) error {
	partCtx, watchdog := newIdleWatchdog(ctx, options.IdleTimeout)
	defer watchdog.stop()
	rawURL, err := buildDownloadURL(options, part, parts)
	if err != nil {
		return err
	}
	response, err := requestDownload(partCtx, client, rawURL, options.Authorization)
	if err != nil {
		if cause := context.Cause(partCtx); cause != nil {
			err = cause
		}
		return fmt.Errorf("download SABR part %d/%d: %w", part+1, parts, err)
	}
	defer response.Body.Close()
	tracks, err := openPartTracks(options, part)
	if err != nil {
		return err
	}
	defer closeTracks(tracks)
	reader := activityReader{reader: response.Body, touch: watchdog.touch}
	if err := consumeDownloadStream(reader, tracks, progress, part == 0); err != nil {
		if cause := context.Cause(partCtx); cause != nil {
			err = cause
		}
		return fmt.Errorf("consume SABR part %d/%d: %w", part+1, parts, err)
	}
	return closeTracks(tracks)
}
