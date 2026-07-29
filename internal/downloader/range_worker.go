package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const defaultCopyBufferSize = 256 * 1024

var copyBufferPool = sync.Pool{
	New: func() any {
		return make([]byte, defaultCopyBufferSize)
	},
}

type chunk struct {
	start int64
	end   int64
}

type progressTracker struct {
	source     Source
	interval   int64
	progress   ProgressFunc
	started    time.Time
	downloaded atomic.Int64
	reportMu   sync.Mutex
	reported   int64
}

func newProgressTracker(source Source, interval int64, progress ProgressFunc) *progressTracker {
	return &progressTracker{
		source:   source,
		interval: interval,
		progress: progress,
		started:  time.Now(),
	}
}

func (tracker *progressTracker) add(bytes int64, now time.Time) {
	current := tracker.downloaded.Add(bytes)
	if tracker.progress == nil || current >= tracker.source.Size {
		return
	}
	tracker.reportMu.Lock()
	defer tracker.reportMu.Unlock()
	if current <= tracker.reported || current-tracker.reported < tracker.interval {
		return
	}
	tracker.reported = current
	tracker.report(current, now)
}

func (tracker *progressTracker) finish(now time.Time) {
	if tracker.progress == nil {
		return
	}
	tracker.reportMu.Lock()
	defer tracker.reportMu.Unlock()
	current := tracker.downloaded.Load()
	if current <= tracker.reported {
		return
	}
	tracker.reported = current
	tracker.report(current, now)
}

func (tracker *progressTracker) report(downloaded int64, now time.Time) {
	elapsed := now.Sub(tracker.started).Seconds()
	if elapsed <= 0 {
		return
	}
	tracker.progress(Progress{
		Name:       tracker.source.Name,
		Downloaded: downloaded,
		Total:      tracker.source.Size,
		Speed:      float64(downloaded) / elapsed,
	})
}

func borrowCopyBuffer(size int) []byte {
	if size == defaultCopyBufferSize {
		return copyBufferPool.Get().([]byte)
	}
	return make([]byte, size)
}

func releaseCopyBuffer(buffer []byte) {
	if cap(buffer) == defaultCopyBufferSize {
		copyBufferPool.Put(buffer[:defaultCopyBufferSize])
	}
}

func downloadChunk(
	ctx context.Context,
	client *http.Client,
	file *os.File,
	source Source,
	part chunk,
	options Options,
	buffer []byte,
	tracker *progressTracker,
) error {
	var lastErr error
	for attempt := 0; attempt < options.Retries; attempt++ {
		if err := waitForRetry(ctx, attempt); err != nil {
			return err
		}
		if err := fetchChunk(ctx, client, file, source, part, options.RangeMode, buffer); err != nil {
			lastErr = err
			continue
		}
		tracker.add(part.end-part.start+1, time.Now())
		return nil
	}
	return fmt.Errorf("%s bytes %d-%d failed: %w", source.Name, part.start, part.end, lastErr)
}

func waitForRetry(ctx context.Context, attempt int) error {
	if attempt == 0 {
		return nil
	}
	timer := time.NewTimer(time.Duration(attempt) * 300 * time.Millisecond)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func copyChunk(body io.Reader, file *os.File, part chunk, buffer []byte) error {
	position := part.start
	for position <= part.end {
		size := min(int64(len(buffer)), part.end-position+1)
		n, err := io.ReadFull(body, buffer[:size])
		if n > 0 {
			written, writeErr := file.WriteAt(buffer[:n], position)
			if writeErr != nil {
				return writeErr
			}
			if written != n {
				return io.ErrShortWrite
			}
			position += int64(n)
		}
		if err != nil {
			return fmt.Errorf("short chunk: got %d expected %d: %w", position-part.start, part.end-part.start+1, err)
		}
	}
	return nil
}
