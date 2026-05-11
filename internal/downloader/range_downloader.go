package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Source struct {
	Name string
	URL  string
	Size int64
}

type Options struct {
	ChunkSize     int64
	Workers       int
	Retries       int
	BufferSize    int
	RangeMode     string
	ProgressBytes int64
}

type Progress struct {
	Name       string
	Downloaded int64
	Total      int64
	Speed      float64
}

type ProgressFunc func(Progress)

func DownloadFile(ctx context.Context, client *http.Client, source Source, output string, options Options, progress ProgressFunc) error {
	if source.URL == "" || source.Size <= 0 {
		return fmt.Errorf("invalid source %s", source.Name)
	}
	if client == nil {
		client = http.DefaultClient
	}
	if options.ChunkSize <= 0 {
		options.ChunkSize = 10 << 20
	}
	if options.Workers <= 0 {
		options.Workers = 8
	}
	if options.Retries <= 0 {
		options.Retries = 4
	}
	if options.BufferSize <= 0 {
		options.BufferSize = 256 * 1024
	}
	if options.ProgressBytes <= 0 {
		options.ProgressBytes = 4 << 20
	}
	if options.RangeMode == "" {
		options.RangeMode = "header"
	}

	tmp := output + ".part"
	_ = os.Remove(tmp)
	file, err := os.OpenFile(tmp, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := file.Truncate(source.Size); err != nil {
		return err
	}

	downloadCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan chunk)
	errs := make(chan error, 1)
	var downloaded atomic.Int64
	started := time.Now()

	var wg sync.WaitGroup
	for range options.Workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				if err := downloadChunk(downloadCtx, client, file, source, job, options, &downloaded, started, progress); err != nil {
					select {
					case errs <- err:
					default:
					}
					cancel()
					return
				}
			}
		}()
	}

sendLoop:
	for start := int64(0); start < source.Size; start += options.ChunkSize {
		end := start + options.ChunkSize - 1
		if end >= source.Size {
			end = source.Size - 1
		}
		select {
		case <-downloadCtx.Done():
			break sendLoop
		case jobs <- chunk{start: start, end: end}:
		}
	}
	close(jobs)
	wg.Wait()

	select {
	case err := <-errs:
		return err
	default:
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if got := downloaded.Load(); got != source.Size {
		return fmt.Errorf("downloaded %d bytes, expected %d", got, source.Size)
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, output)
}

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

func fetchChunk(ctx context.Context, client *http.Client, file *os.File, source Source, part chunk, options Options, downloaded *atomic.Int64, started time.Time, progress ProgressFunc) error {
	requestURL, err := rangedURL(source.URL, part, options.RangeMode)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept-Encoding", "identity")
	if options.RangeMode == "header" {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", part.start, part.end))
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/122 Safari/537.36")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusPartialContent && !(options.RangeMode == "query" && res.StatusCode == http.StatusOK) {
		return fmt.Errorf("unexpected HTTP status %d", res.StatusCode)
	}
	if contentRange := res.Header.Get("Content-Range"); options.RangeMode == "header" && !strings.HasPrefix(contentRange, fmt.Sprintf("bytes %d-%d/", part.start, part.end)) {
		return fmt.Errorf("unexpected Content-Range %q", contentRange)
	}

	buf := make([]byte, options.BufferSize)
	position := part.start
	lastProgress := downloaded.Load()
	for {
		n, readErr := res.Body.Read(buf)
		if n > 0 {
			if position+int64(n)-1 > part.end {
				return fmt.Errorf("chunk overflow")
			}
			if _, err := file.WriteAt(buf[:n], position); err != nil {
				return err
			}
			position += int64(n)
			current := downloaded.Add(int64(n))
			if progress != nil && (current-lastProgress >= options.ProgressBytes || current == source.Size) {
				lastProgress = current
				elapsed := time.Since(started).Seconds()
				if elapsed > 0 {
					progress(Progress{Name: source.Name, Downloaded: current, Total: source.Size, Speed: float64(current) / elapsed})
				}
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	if position != part.end+1 {
		return fmt.Errorf("short chunk: got %d expected %d", position-part.start, part.end-part.start+1)
	}
	return nil
}

func rangedURL(rawURL string, part chunk, mode string) (string, error) {
	if mode == "header" {
		return rawURL, nil
	}
	if mode != "query" {
		return "", fmt.Errorf("unsupported range mode %q", mode)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	query := parsed.Query()
	query.Set("range", fmt.Sprintf("%d-%d", part.start, part.end))
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}
