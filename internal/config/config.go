package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr         string
	PublicBaseURL    string
	TypeTypeAPIBase  string
	DataDir          string
	DatabaseURL      string
	RedisAddr        string
	JobTTLSeconds    int
	StorageBackend   string
	S3Endpoint       string
	S3PublicEndpoint string
	S3Region         string
	S3Bucket         string
	S3AccessKey      string
	S3SecretKey      string
	S3UseSSL         bool
	S3PublicUseSSL   bool
	S3PathStyle      bool
	ArtifactTTL      int
	MaxWorkers       int
	DownloadWorkers  int
	ChunkSize        int64
	RangeMode        string
	Muxer            string
	HTTP2            bool
	MaxQueueSize     int
}

func Load() Config {
	s3Endpoint, s3SSL := parseS3Endpoint("S3_ENDPOINT")
	s3PublicEndpoint, s3PublicSSL := parseS3Endpoint("S3_PUBLIC_ENDPOINT")
	return Config{
		HTTPAddr:         httpAddr(),
		PublicBaseURL:    strings.TrimRight(env("PUBLIC_BASE_URL", "http://localhost:18093"), "/"),
		TypeTypeAPIBase:  strings.TrimRight(env("TYPETYPE_API_BASE", "http://typetype-server:8080"), "/"),
		DataDir:          env("DATA_DIR", "data"),
		DatabaseURL:      databaseURL(),
		RedisAddr:        redisAddr(),
		JobTTLSeconds:    envInt("JOB_TTL_SECONDS", 600),
		StorageBackend:   storageBackend(),
		S3Endpoint:       s3Endpoint,
		S3PublicEndpoint: s3PublicEndpoint,
		S3Region:         env("S3_REGION", "garage"),
		S3Bucket:         env("S3_BUCKET", ""),
		S3AccessKey:      env("S3_ACCESS_KEY", ""),
		S3SecretKey:      env("S3_SECRET_KEY", ""),
		S3UseSSL:         envBool("S3_USE_SSL", s3SSL),
		S3PublicUseSSL:   envBool("S3_PUBLIC_USE_SSL", s3PublicSSL),
		S3PathStyle:      envBool("S3_PATH_STYLE", true),
		ArtifactTTL:      envInt("ARTIFACT_URL_TTL_SECONDS", envInt("S3_ARTIFACT_TTL_SECONDS", 7200)),
		MaxWorkers:       envInt("MAX_CONCURRENT_WORKERS", 2),
		DownloadWorkers:  envInt("DOWNLOAD_WORKERS", 8),
		ChunkSize:        envSize("DOWNLOAD_CHUNK_SIZE", 10<<20),
		RangeMode:        env("DOWNLOAD_RANGE_MODE", "query"),
		Muxer:            env("MUXER", "avformat"),
		HTTP2:            envBool("DOWNLOAD_HTTP2", true),
		MaxQueueSize:     envInt("MAX_QUEUE_SIZE", 100),
	}
}

func redisAddr() string {
	if value := env("REDIS_ADDR", ""); value != "" {
		return value
	}
	host := env("REDIS_HOST", "")
	if host == "" {
		return ""
	}
	return host + ":" + env("REDIS_PORT", "6379")
}

func httpAddr() string {
	if value := env("HTTP_ADDR", ""); value != "" {
		return value
	}
	return ":" + env("HTTP_PORT", "18093")
}

func databaseURL() string {
	if value := env("DATABASE_URL", ""); value != "" {
		return value
	}
	raw := env("DB_URL", "")
	if raw == "" {
		return ""
	}
	user := url.QueryEscape(env("DB_USER", "typetype"))
	password := url.QueryEscape(env("DB_PASSWORD", "typetype"))
	if strings.HasPrefix(raw, "jdbc:postgresql://") {
		return "postgres://" + user + ":" + password + "@" + strings.TrimPrefix(raw, "jdbc:postgresql://") + "?sslmode=disable"
	}
	return raw
}

func storageBackend() string {
	if value := strings.ToLower(env("STORAGE_BACKEND", "")); value != "" {
		return value
	}
	if env("S3_ENDPOINT", "") != "" && env("S3_BUCKET", "") != "" && env("S3_ACCESS_KEY", "") != "" {
		return "s3"
	}
	return "local"
}

func parseS3Endpoint(name string) (string, bool) {
	raw := env(name, "")
	if raw == "" {
		return "", true
	}
	parsed, err := url.Parse(raw)
	if err == nil && parsed.Host != "" {
		return parsed.Host, parsed.Scheme == "https"
	}
	return strings.TrimPrefix(strings.TrimPrefix(raw, "https://"), "http://"), true
}

func env(name string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(name string, fallback int) int {
	value, err := strconv.Atoi(env(name, ""))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func envBool(name string, fallback bool) bool {
	value := strings.ToLower(env(name, ""))
	if value == "true" || value == "1" || value == "yes" {
		return true
	}
	if value == "false" || value == "0" || value == "no" {
		return false
	}
	return fallback
}

func envSize(name string, fallback int64) int64 {
	value, err := strconv.ParseInt(env(name, ""), 10, 64)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func ComposePostgresURL(host string, port int, database string, user string, password string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", url.QueryEscape(user), url.QueryEscape(password), host, port, database)
}
