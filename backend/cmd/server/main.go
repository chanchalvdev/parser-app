package main

import (
\t"context"
\t"log"

\t"github.com/example/file-platform/backend/internal/config"
\t"github.com/example/file-platform/backend/internal/db"
\t"github.com/example/file-platform/backend/internal/queue"
\t"github.com/example/file-platform/backend/internal/repo"
\t"github.com/example/file-platform/backend/internal/search"
\t"github.com/example/file-platform/backend/internal/storage"
\thandlers "github.com/example/file-platform/backend/internal/http"
\t"github.com/joho/godotenv"
)

func main() {
\t_ = godotenv.Load()
\tcfg, err := config.Load()
\tif err != nil {
\t\tlog.Fatalf("config: %v", err)
\t}
\tctx := context.Background()
\tpool, err := db.NewPool(ctx, cfg.DatabaseURL)
\tif err != nil {
\t\tlog.Fatalf("db: %v", err)
\t}
\tdefer pool.Close()
\tif err := db.RunMigrations(ctx, pool, "migrations"); err != nil {
\t\tlog.Fatalf("migrations: %v", err)
\t}
\trepository := repo.New(pool)
\tif err := repository.DB.QueryRow(ctx, "SELECT 1").Scan(new(int)); err != nil {
\t\tlog.Fatalf("db readiness: %v", err)
\t}
\tminioClient, err := storage.New(cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, cfg.MinioUseSSL)
\tif err != nil {
\t\tlog.Fatalf("minio: %v", err)
\t}
\tif err := minioClient.EnsureBucket(ctx, cfg.MinioBucket); err != nil {
\t\tlog.Fatalf("minio bucket: %v", err)
\t}
\tq := queue.NewProducer(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB, cfg.QueueName)
\tsrch, err := search.New(search.Config{
\t\tURL:      cfg.OpenSearchURL,
\t\tUser:     cfg.OpenSearchUser,
\t\tPassword: cfg.OpenSearchPass,
\t\tIndex:    cfg.OpenSearchIndex,
\t})
\tif err != nil {
\t\tlog.Printf("opensearch unavailable: %v", err)
\t\tsrch = nil
\t}
\trouter := handlers.New(cfg, repository, q, minioClient, srch).Router()
\tif err := router.Run(cfg.APIListenAddr); err != nil {
\t\tlog.Fatalf("listen: %v", err)
\t}
}
