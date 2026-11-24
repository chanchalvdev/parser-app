package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/enterprise-file-platform/api/internal/config"
	"github.com/enterprise-file-platform/api/internal/db"
	"github.com/enterprise-file-platform/api/internal/http/handlers"
	"github.com/enterprise-file-platform/api/internal/http/middleware"
	"github.com/enterprise-file-platform/api/internal/http/routes"
	"github.com/enterprise-file-platform/api/internal/observability"
	"github.com/enterprise-file-platform/api/internal/queue"
	"github.com/enterprise-file-platform/api/internal/search"
	"github.com/enterprise-file-platform/api/internal/repositories"
	"github.com/enterprise-file-platform/api/internal/services"
	"github.com/enterprise-file-platform/api/internal/storage"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("load config: %v", err))
	}

	logger, err := observability.NewLogger(cfg.AppEnv)
	if err != nil {
		panic(fmt.Sprintf("init logger: %v", err))
	}
	defer logger.Sync()

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("postgres unavailable at startup", zap.Error(err))
	}
	defer pool.Close()

	objectStore, err := storage.New(storage.Config{
		Backend:              cfg.StorageBackend,
		Bucket:               cfg.MinioBucket,
		MinioEndpoint:        cfg.MinioEndpoint,
		MinioPresignEndpoint: cfg.MinioPresignEndpoint,
		MinioAccessKey:       cfg.MinioAccessKey,
		MinioSecretKey:       cfg.MinioSecretKey,
		AzureAccountName:     cfg.AzureStorageAccount,
		AzureAccountKey:      cfg.AzureStorageKey,
		AzureEndpoint:        cfg.AzureStorageEndpoint,
		AzurePublicEndpoint:  cfg.AzureStoragePublicEndpoint,
	})
	if err != nil {
		logger.Fatal("init object storage failed", zap.Error(err))
	}
	if err := objectStore.EnsureBucket(ctx); err != nil {
		logger.Fatal("ensure storage bucket failed", zap.String("backend", objectStore.Backend()), zap.Error(err))
	}
	logger.Info("object storage ready", zap.String("backend", objectStore.Backend()), zap.String("bucket", objectStore.Bucket()))
	redisOptions := &redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}
	redisClient := redis.NewClient(redisOptions)
	defer redisClient.Close()
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logger.Fatal("redis unavailable at startup", zap.Error(err))
	}
	queueProducer := queue.NewRedisProducer(redisClient, cfg.RedisQueueName)

	healthRepo := repositories.NewPostgresHealthRepository(pool)
	healthService := services.NewHealthService(healthRepo)
	versionService := services.NewVersionService()
	uploadRepo := repositories.NewUploadRepository(pool)
	fileRepo := repositories.NewFileRepository(pool)
	parsedRecordRepo := repositories.NewParsedRecordRepository(pool)
	archivePasswordRefRepo := repositories.NewArchivePasswordRefRepository(pool)
	dashboardRepo := repositories.NewDashboardRepository(pool)
	auditLogRepo := repositories.NewAuditLogRepository(pool)
	jobRepo := repositories.NewJobRepository(pool)
	jobEventRepo := repositories.NewJobEventRepository(pool)
	settingsRepo := repositories.NewSettingsRepository(pool)
	uploadService := services.NewUploadService(
		uploadRepo,
		fileRepo,
		jobRepo,
		jobEventRepo,
		settingsRepo,
		auditLogRepo,
		objectStore,
		queueProducer,
	)
	fileService := services.NewFileService(
		fileRepo,
		parsedRecordRepo,
		jobRepo,
		jobEventRepo,
		archivePasswordRefRepo,
		auditLogRepo,
		queueProducer,
		cfg.MinioBucket,
	)
	adminSettingsService := services.NewAdminSettingsService(settingsRepo, auditLogRepo)
	jobService := services.NewJobService(
		jobRepo,
		jobEventRepo,
		fileRepo,
		parsedRecordRepo,
		auditLogRepo,
		queueProducer,
		cfg.MinioBucket,
	)
	dashboardService := services.NewDashboardService(dashboardRepo)
	searchClient := search.NewClient(search.ClientConfig{
		URL:                 cfg.OpenSearchURL,
		Username:            cfg.OpenSearchUser,
		Password:            cfg.OpenSearchPassword,
		ParsedRecordsIndex:   cfg.OpenSearchParsedRecordsIndex,
		FilesIndex:          cfg.OpenSearchFilesIndex,
	})
	searchService := services.NewSearchService(searchClient, auditLogRepo)
	searchHandler := handlers.NewSearchHandler(searchService)

	handler := handlers.NewHealthHandler(healthService)
	versionHandler := handlers.NewVersionHandler(versionService, cfg.AppEnv)
	docsHandler := handlers.NewDocsHandler()
	uploadHandler := handlers.NewUploadHandler(uploadService)
	fileHandler := handlers.NewFileHandler(fileService)
	jobHandler := handlers.NewJobHandler(jobService)
	adminSettingsHandler := handlers.NewAdminSettingsHandler(adminSettingsService)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)

	router := routes.NewRouter(
		handler,
		versionHandler,
		docsHandler,
		uploadHandler,
		searchHandler,
		fileHandler,
		jobHandler,
		dashboardHandler,
		adminSettingsHandler,
		middleware.RequestID,
		middleware.SetAuthContext(cfg.AuthEnabled),
		func(next http.Handler) http.Handler {
			return middleware.CORS(cfg.AllowedOrigins)(next)
		},
		func(next http.Handler) http.Handler {
			return middleware.Recovery(logger)(next)
		},
		middleware.Logger(logger),
		// With AUTH_ENABLED=true every /api/v1 call needs X-API-Key (Render/Azure).
		middleware.RequireAPIKey(apiKeyIfAuthEnabled(cfg)),
	)

	address := fmt.Sprintf(":%d", cfg.APIPort)
	srv := &http.Server{
		Addr:              address,
		Handler:           router.Handler(),
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("starting API", zap.String("addr", address))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failed", zap.Error(err))
			os.Exit(1)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("shutdown failed", zap.Error(err))
	}
}

func apiKeyIfAuthEnabled(cfg *config.Config) string {
	if !cfg.AuthEnabled {
		return ""
	}
	return cfg.APIKey
}
