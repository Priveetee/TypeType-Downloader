package config

import "testing"

func TestLoadReadsStorageAndDatabaseConfig(t *testing.T) {
	t.Setenv("HTTP_PORT", "18093")
	t.Setenv("DB_URL", "jdbc:postgresql://postgres:5432/typetype_downloader")
	t.Setenv("DB_USER", "typetype")
	t.Setenv("DB_PASSWORD", "typetype")
	t.Setenv("REDIS_HOST", "dragonfly")
	t.Setenv("REDIS_PORT", "6379")
	t.Setenv("STORAGE_BACKEND", "s3")
	t.Setenv("S3_ENDPOINT", "http://garage:3900")
	t.Setenv("S3_PUBLIC_ENDPOINT", "http://localhost:3900")
	t.Setenv("S3_BUCKET", "downloads")
	t.Setenv("S3_ACCESS_KEY", "key")
	t.Setenv("S3_SECRET_KEY", "secret")
	t.Setenv("S3_ARTIFACT_TTL_SECONDS", "7200")
	cfg := Load()
	if cfg.HTTPAddr != ":18093" || cfg.StorageBackend != "s3" || cfg.S3UseSSL {
		t.Fatalf("unexpected config: %#v", cfg)
	}
	if cfg.DatabaseURL != "postgres://typetype:typetype@postgres:5432/typetype_downloader?sslmode=disable" {
		t.Fatalf("database URL = %q", cfg.DatabaseURL)
	}
	if cfg.RedisAddr != "dragonfly:6379" {
		t.Fatalf("redis addr = %q", cfg.RedisAddr)
	}
	if cfg.S3Endpoint != "garage:3900" || cfg.S3PublicEndpoint != "localhost:3900" || cfg.S3Bucket != "downloads" || cfg.ArtifactTTL != 7200 {
		t.Fatalf("unexpected S3 config: %#v", cfg)
	}
	if cfg.S3PublicUseSSL {
		t.Fatalf("unexpected public S3 SSL config: %#v", cfg)
	}
}
