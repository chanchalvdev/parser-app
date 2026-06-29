package config

import (
\t"os"
\t"strconv"
)

type Config struct {
\tAPIListenAddr    string
\tDatabaseURL      string
\tRedisAddr        string
\tRedisPassword    string
\tRedisDB          int
\tQueueName        string
\tMinioEndpoint    string
\tMinioAccessKey   string
\tMinioSecretKey   string
\tMinioBucket      string
\tMinioUseSSL      bool
\tOpenSearchURL    string
\tOpenSearchUser   string
\tOpenSearchPass   string
\tOpenSearchIndex  string
\tAPIKey           string
\tAuthEnabled      bool
\tMaxUploadSizeMB  int
\tUploadPathPrefix string
}

func getenv(key, def string) string {
\tif v := os.Getenv(key); v != "" {
\t\treturn v
\t}
\treturn def
}

func getenvInt(key string, def int) int {
\tv := getenv(key, "")
\tif v == "" {
\t\treturn def
\t}
\tn, err := strconv.Atoi(v)
\tif err != nil {
\t\treturn def
\t}
\treturn n
}

func Load() (Config, error) {
\tcfg := Config{
\t\tAPIListenAddr:    getenv("API_LISTEN_ADDR", ":8080"),
\t\tDatabaseURL:      getenv("DATABASE_URL", "postgres://postgres:postgres@postgres:5432/fileplatform?sslmode=disable"),
\t\tRedisAddr:        getenv("REDIS_ADDR", "redis:6379"),
\t\tRedisPassword:    getenv("REDIS_PASSWORD", ""),
\t\tRedisDB:          getenvInt("REDIS_DB", 0),
\t\tQueueName:        getenv("QUEUE_NAME", "ingest.jobs"),
\t\tMinioEndpoint:    getenv("MINIO_ENDPOINT", "minio:9000"),
\t\tMinioAccessKey:   getenv("MINIO_ACCESS_KEY", "minioadmin"),
\t\tMinioSecretKey:   getenv("MINIO_SECRET_KEY", "minioadmin"),
\t\tMinioBucket:      getenv("MINIO_BUCKET", "ingest"),
\t\tMinioUseSSL:      getenv("MINIO_USE_SSL", "false") == "true",
\t\tOpenSearchURL:    getenv("OPENSEARCH_URL", "http://opensearch:9200"),
\t\tOpenSearchUser:   getenv("OPENSEARCH_USER", "admin"),
\t\tOpenSearchPass:   getenv("OPENSEARCH_PASSWORD", "admin"),
\t\tOpenSearchIndex:  getenv("OPENSEARCH_INDEX", "parsed-documents"),
\t\tAPIKey:           getenv("API_KEY", ""),
\t\tAuthEnabled:      getenv("AUTH_ENABLED", "false") == "true",
\t\tMaxUploadSizeMB:  getenvInt("MAX_UPLOAD_SIZE_MB", 250),
\t\tUploadPathPrefix: getenv("UPLOAD_PATH_PREFIX", "uploads"),
\t}
\treturn cfg, nil
}
