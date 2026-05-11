package db

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"typetype-downloader-go/internal/job"
)

type DragonflySink struct {
	client *redis.Client
	ttl    time.Duration
	queue  chan *job.Record
}

func OpenDragonfly(addr string, ttlSeconds int) *DragonflySink {
	if ttlSeconds <= 0 {
		ttlSeconds = 600
	}
	sink := &DragonflySink{
		client: redis.NewClient(&redis.Options{Addr: addr}),
		ttl:    time.Duration(ttlSeconds) * time.Second,
		queue:  make(chan *job.Record, 2048),
	}
	go sink.run()
	return sink
}

func (s *DragonflySink) Close() error { return s.client.Close() }

func (s *DragonflySink) Name() string { return "dragonfly" }

func (s *DragonflySink) Health(ctx context.Context) error { return s.client.Ping(ctx).Err() }

func (s *DragonflySink) SaveJob(record *job.Record) {
	select {
	case s.queue <- record:
	default:
	}
}

func (s *DragonflySink) run() {
	for record := range s.queue {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_ = s.save(ctx, record)
		cancel()
	}
}

func (s *DragonflySink) save(ctx context.Context, record *job.Record) error {
	payload, err := json.Marshal(record)
	if err != nil {
		return err
	}
	pipe := s.client.Pipeline()
	pipe.Set(ctx, "downloader:job:"+record.ID, payload, s.ttl)
	pipe.Set(ctx, "downloader:status:"+record.ID, string(record.Status), s.ttl)
	if record.CacheKey != "" && record.Status == job.StatusDone {
		pipe.Set(ctx, "downloader:cache:"+record.CacheKey, record.ID, s.ttl)
	}
	_, err = pipe.Exec(ctx)
	return err
}
