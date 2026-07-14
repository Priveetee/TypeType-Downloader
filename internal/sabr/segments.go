package sabr

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

func downloadFiles(ctx context.Context, client *http.Client, files []filePlan, authorization string, workers int, progress *reporter) error {
	if workers < 1 {
		workers = 1
	}
	if workers > 4 {
		workers = 4
	}
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	queue := make(chan filePlan, len(files))
	for _, file := range files {
		queue <- file
	}
	close(queue)
	errs := make(chan error, 1)
	var group sync.WaitGroup
	for range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			for file := range queue {
				if err := downloadFile(workerCtx, client, file, authorization); err != nil {
					select {
					case errs <- err:
						cancel()
					default:
					}
					return
				}
				info, err := os.Stat(file.Path)
				if err != nil {
					select {
					case errs <- err:
						cancel()
					default:
					}
					return
				}
				progress.add(info.Size())
			}
		}()
	}
	group.Wait()
	select {
	case err := <-errs:
		return err
	default:
		return ctx.Err()
	}
}

func downloadFile(ctx context.Context, client *http.Client, file filePlan, authorization string) error {
	var last error
	for attempt := 1; attempt <= 4; attempt++ {
		response, err := request(ctx, client, file.URL, authorization)
		if err == nil {
			err = writeResponse(file.Path, response)
		}
		if err == nil {
			return nil
		}
		last = err
		if attempt < 4 {
			if err := retryDelay(ctx, attempt); err != nil {
				return err
			}
		}
	}
	return fmt.Errorf("download SABR segment failed after 4 attempts: %w", last)
}

func retryDelay(ctx context.Context, attempt int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Duration(attempt) * 250 * time.Millisecond):
		return nil
	}
}

func writeResponse(path string, response *http.Response) error {
	defer response.Body.Close()
	tempPath := path + ".download"
	output, err := os.Create(tempPath)
	if err != nil {
		return err
	}
	written, copyErr := io.Copy(output, response.Body)
	closeErr := output.Close()
	if copyErr != nil || closeErr != nil || written == 0 {
		_ = os.Remove(tempPath)
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		return fmt.Errorf("empty SABR segment")
	}
	return os.Rename(tempPath, path)
}

func assemble(ctx context.Context, target string, parts []string) error {
	output, err := os.Create(target)
	if err != nil {
		return err
	}
	completed := false
	defer func() {
		if !completed {
			output.Close()
			_ = os.Remove(target)
		}
	}()
	for _, path := range parts {
		if err := ctx.Err(); err != nil {
			return err
		}
		input, openErr := os.Open(path)
		if openErr != nil {
			return openErr
		}
		_, copyErr := io.Copy(output, input)
		closeErr := input.Close()
		if copyErr != nil || closeErr != nil {
			if copyErr != nil {
				return copyErr
			}
			return closeErr
		}
		_ = os.Remove(path)
	}
	if err := output.Close(); err != nil {
		return err
	}
	completed = true
	return nil
}
