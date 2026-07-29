package sabr

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync/atomic"
	"time"
)

var errDownloadIdleTimeout = errors.New("SABR download stream timed out")

type idleWatchdog struct {
	cancel  context.CancelCauseFunc
	last    atomic.Int64
	timeout time.Duration
}

func newIdleWatchdog(parent context.Context, timeout time.Duration) (context.Context, *idleWatchdog) {
	if timeout <= 0 {
		timeout = defaultDownloadIdleTimeout
	}
	ctx, cancel := context.WithCancelCause(parent)
	watchdog := &idleWatchdog{cancel: cancel, timeout: timeout}
	watchdog.touch()
	go watchdog.run(ctx)
	return ctx, watchdog
}

func (w *idleWatchdog) run(ctx context.Context) {
	timer := time.NewTimer(w.timeout)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-timer.C:
			inactive := now.Sub(time.Unix(0, w.last.Load()))
			if inactive >= w.timeout {
				w.cancel(fmt.Errorf("%w after %s", errDownloadIdleTimeout, w.timeout))
				return
			}
			timer.Reset(w.timeout - inactive)
		}
	}
}

func (w *idleWatchdog) touch() {
	w.last.Store(time.Now().UnixNano())
}

func (w *idleWatchdog) stop() {
	w.cancel(nil)
}

type activityReader struct {
	reader io.Reader
	touch  func()
}

func (r activityReader) Read(buffer []byte) (int, error) {
	count, err := r.reader.Read(buffer)
	if count > 0 {
		r.touch()
	}
	return count, err
}

const defaultDownloadIdleTimeout = 60 * time.Second
