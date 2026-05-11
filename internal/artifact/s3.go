package artifact

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Config struct {
	Endpoint       string
	PublicEndpoint string
	Region         string
	Bucket         string
	AccessKey      string
	SecretKey      string
	UseSSL         bool
	PublicUseSSL   bool
	PathStyle      bool
	URLTTL         time.Duration
}

type S3Store struct {
	client  *minio.Client
	presign *minio.Client
	cfg     S3Config
}

func (s *S3Store) Name() string { return "s3" }

func (s *S3Store) Health(ctx context.Context) error {
	_, err := s.client.BucketExists(ctx, s.cfg.Bucket)
	return err
}

func NewS3Store(cfg S3Config) (*S3Store, error) {
	if cfg.Endpoint == "" || cfg.Bucket == "" || cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("S3 endpoint, bucket, access key, and secret key are required")
	}
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure:       cfg.UseSSL,
		Region:       cfg.Region,
		BucketLookup: bucketLookup(cfg.PathStyle),
	})
	if err != nil {
		return nil, err
	}
	presign := client
	if cfg.PublicEndpoint != "" {
		presign, err = minio.New(cfg.PublicEndpoint, &minio.Options{
			Creds:        credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
			Secure:       cfg.PublicUseSSL,
			Region:       cfg.Region,
			BucketLookup: bucketLookup(cfg.PathStyle),
		})
		if err != nil {
			return nil, err
		}
	}
	if cfg.URLTTL <= 0 {
		cfg.URLTTL = 15 * time.Minute
	}
	return &S3Store{client: client, presign: presign, cfg: cfg}, nil
}

func (s *S3Store) Save(ctx context.Context, localPath string, objectKey string) (Saved, error) {
	file, err := os.Open(localPath)
	if err != nil {
		return Saved{}, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return Saved{}, err
	}
	contentType := mime.TypeByExtension(filepath.Ext(localPath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	_, err = s.client.PutObject(ctx, s.cfg.Bucket, objectKey, file, info.Size(), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return Saved{}, err
	}
	return Saved{Backend: "s3", Location: objectKey, Expires: time.Now().Add(s.cfg.URLTTL)}, nil
}

func (s *S3Store) ServeHTTP(w http.ResponseWriter, r *http.Request, saved Saved, fileName string) error {
	values := make(map[string][]string)
	if fileName != "" {
		values["response-content-disposition"] = []string{`attachment; filename="` + fileName + `"`}
	}
	url, err := s.presign.PresignedGetObject(r.Context(), s.cfg.Bucket, saved.Location, s.cfg.URLTTL, values)
	if err != nil {
		return err
	}
	http.Redirect(w, r, url.String(), http.StatusFound)
	return nil
}

func (s *S3Store) Delete(ctx context.Context, saved Saved) error {
	if saved.Location == "" {
		return nil
	}
	return s.client.RemoveObject(ctx, s.cfg.Bucket, saved.Location, minio.RemoveObjectOptions{})
}

func bucketLookup(pathStyle bool) minio.BucketLookupType {
	if pathStyle {
		return minio.BucketLookupPath
	}
	return minio.BucketLookupDNS
}
