package db

import (
	"context"
	"encoding/json"
	"time"

	"typetype-downloader-go/internal/job"
)

func (s *PostgresSink) LoadDone(ctx context.Context) ([]*job.Record, error) {
	return s.load(ctx, loadDoneSQL)
}

func (s *PostgresSink) LoadRunnable(ctx context.Context) ([]*job.Record, error) {
	return s.load(ctx, loadRunnableSQL)
}

func (s *PostgresSink) load(ctx context.Context, sql string) ([]*job.Record, error) {
	rows, err := s.pool.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []*job.Record
	for rows.Next() {
		record, err := scanRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRecord(row rowScanner) (*job.Record, error) {
	var record job.Record
	var status string
	var optionsRaw, progressRaw, resolvedRaw []byte
	var queuedAt time.Time
	err := row.Scan(
		&record.ID, &record.CacheKey, &record.URL, &status, &record.Title,
		&optionsRaw, &progressRaw, &resolvedRaw, &record.Artifact, &record.Storage,
		&record.ExpiresAt, &record.Error, &record.ErrorCode, &queuedAt, &record.StartedAt, &record.FinishedAt,
		&record.DownloadMs, &record.MuxMs, &record.TotalMs,
	)
	if err != nil {
		return nil, err
	}
	record.Status = job.Status(status)
	record.QueuedAt = queuedAt
	if err := json.Unmarshal(optionsRaw, &record.Options); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(progressRaw, &record.Progress); err != nil {
		return nil, err
	}
	if len(resolvedRaw) > 0 && string(resolvedRaw) != "null" {
		var resolved job.ResolvedOutput
		if err := json.Unmarshal(resolvedRaw, &resolved); err != nil {
			return nil, err
		}
		record.Resolved = &resolved
	}
	return &record, nil
}
