package sabr

import (
	"sync"
	"time"
)

type reporter struct {
	mu         sync.Mutex
	downloaded int64
	last       time.Time
	progress   ProgressFunc
}

func newReporter(progress ProgressFunc) *reporter {
	return &reporter{last: time.Now(), progress: progress}
}

func (r *reporter) add(bytes int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.downloaded += bytes
	if r.progress != nil && time.Since(r.last) >= 250*time.Millisecond {
		r.last = time.Now()
		r.progress(r.downloaded)
	}
}

func (r *reporter) finish() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.progress != nil {
		r.progress(r.downloaded)
	}
}
