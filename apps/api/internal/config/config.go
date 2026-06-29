package config

import (
	"fmt"
	"os"
	"strings"
	"strconv"
)

type Config struct {
	AppEnv         string
	APIPort        int
	AuthEnabled    bool
	DatabaseURL    string
	RedisAddr      string
	RedisPassword  string
	RedisDB        int
	RedisQueueName string
	MinioEndpoint string
	MinioPresignEndpoint string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	OpenSearchURL  string
	OpenSearchUser string
	OpenSearchPassword string
	OpenSearchParsedRecordsIndex string
	OpenSearchFilesIndex string
	AllowedOrigins []string
}

func Load() (*Config, error) {
	cfg := &Config{
		AppEnv:         getenv("APP_ENV", "local"),
		APIPort:        getenvInt("API_PORT", 8080),
		AuthEnabled:    getenvBool("AUTH_ENABLED", false),
		DatabaseURL:    getenv("DATABASE_URL", "postgres://file_user:file_password@postgres:5432/file_platform?sslmode=disable"),
		RedisAddr:      getenv("REDIS_ADDR", "redis:6379"),
		RedisPassword:  getenv("REDIS_PASSWORD", ""),
		RedisDB:        getenvInt("REDIS_DB", 0),
		RedisQueueName: getenv("QUEUE_NAME", "ingestion_jobs"),
		MinioEndpoint:         getenv("MINIO_ENDPOINT", "minio:9000"),
		MinioPresignEndpoint:  getenv("MINIO_PRESIGN_ENDPOINT", getenv("MINIO_ENDPOINT", "minio:9000")),
		MinioAccessKey: getenv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey: getenv("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:    getenv("MINIO_BUCKET", "file-ingestion"),
		OpenSearchURL:  getenv("OPENSEARCH_URL", "http://opensearch:9200"),
		OpenSearchUser: getenv("OPENSEARCH_USER", ""),
		OpenSearchPassword: getenv("OPENSEARCH_PASSWORD", ""),
		OpenSearchParsedRecordsIndex: getenv("OPENSEARCH_PARSED_RECORDS_INDEX", "parsed-records"),
		OpenSearchFilesIndex: getenv("OPENSEARCH_FILES_INDEX", "files"),
		AllowedOrigins:  parseAllowedOrigins(getenv("ALLOWED_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173")),
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	return cfg, nil
}

func getenv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getenvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func getenvBool(key string, defaultValue bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return defaultValue
	}
	return value == "1" || value == "true" || value == "t" || value == "yes" || value == "y"
}

func parseAllowedOrigins(raw string) []string {
	parts := []string{}
	for _, value := range strings.Split(raw, ",") {
		origin := strings.TrimSpace(value)
		if origin == "" {
			continue
		}
		parts = append(parts, origin)
	}
	return parts
}
