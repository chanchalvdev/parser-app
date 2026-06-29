package handlers

import (
\t"context"
\t"fmt"
\t"log"
\t"net/http"
\t"path/filepath"
\t"strconv"
\t"strings"
\t"time"
\t"sync"

\t"github.com/example/file-platform/backend/internal/config"
\t"github.com/example/file-platform/backend/internal/models"
\t"github.com/example/file-platform/backend/internal/middleware"
\t"github.com/example/file-platform/backend/internal/queue"
\t"github.com/example/file-platform/backend/internal/repo"
\t"github.com/example/file-platform/backend/internal/search"
\t"github.com/example/file-platform/backend/internal/storage"
\t"github.com/gin-gonic/gin"
\t"github.com/google/uuid"
\t"github.com/jackc/pgx/v5"
)

type API struct {
\tcfg         config.Config
\trepo        *repo.Repository
\tqueue       *queue.Producer
\tminio       *storage.Client
\tsearch      *search.Client
\tsettings    map[string]any
\tsettingsMu  sync.RWMutex
}

func New(cfg config.Config, repo *repo.Repository, q *queue.Producer, minioClient *storage.Client, searchClient *search.Client) *API {
\treturn &API{
\t\tcfg:        cfg,
\t\trepo:       repo,
\t\tqueue:      q,
\t\tminio:      minioClient,
\t\tsearch:     searchClient,
\t\tsettings: map[string]any{
\t\t\t"max_upload_size_mb": cfg.MaxUploadSizeMB,
\t\t\t"queue_name":        cfg.QueueName,
\t\t\t"index_name":        cfg.OpenSearchIndex,
\t\t},
\t}
}

func (api *API) Router() *gin.Engine {
\tr := gin.Default()
\tr.Use(middleware.CORS(), middleware.Auth(api.cfg))
\tr.GET("/health", api.health)
\tr.GET("/ready", api.ready)

\tv1 := r.Group("/api/v1")
\t{
\t\tv1.POST("/uploads", api.uploadFile)
\t\tv1.GET("/uploads", api.listUploads)
\t\tv1.GET("/uploads/:id", api.getUpload)
\t\tv1.GET("/jobs", api.listJobs)
\t\tv1.GET("/jobs/:id", api.getJob)
\t\tv1.POST("/jobs/:id/password", api.retryWithPassword)
\t\tv1.GET("/search", api.searchContent)
\t\tv1.GET("/files/:uploadId/tree", api.fileTree)
\t\tv1.GET("/dashboard/summary", api.dashboardSummary)
\t\tv1.GET("/audit", api.listAuditLogs)
\t\tv1.GET("/admin/settings", api.getSettings)
\t\tv1.PUT("/admin/settings", api.updateSettings)
\t}
\treturn r
}

func (api *API) health(c *gin.Context) {
\tc.JSON(http.StatusOK, gin.H{
\t\t"status": "ok",
\t\t"time":   time.Now().UTC().Format(time.RFC3339),
\t})
}

func (api *API) ready(c *gin.Context) {
\tctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
\tdefer cancel()
\tif err := api.repo.DB.Ping(ctx); err != nil {
\t\tc.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "reason": err.Error()})
\t\treturn
\t}
\tc.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func (api *API) uploadFile(c *gin.Context) {
\tuserID := middleware.UserID(c)
\tif err := c.Request.ParseMultipartForm(int64(api.cfg.MaxUploadSizeMB) * 1024 * 1024); err != nil {
\t\tc.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart body"})
\t\treturn
\t}
\tfile, header, err := c.Request.FormFile("file")
\tif err != nil {
\t\tc.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
\t\treturn
\t}
\tdefer file.Close()
\tfilename := strings.TrimSpace(filepath.Base(header.Filename))
\tif filename == "" {
\t\tfilename = "upload.bin"
\t}
\tpassword := strings.TrimSpace(c.PostForm("archive_password"))
\tobjectKey := fmt.Sprintf("%s/%s/%s", api.cfg.UploadPathPrefix, time.Now().UTC().Format("20060102"), uuid.NewString())
\tif err := api.minio.PutObject(context.Background(), api.cfg.MinioBucket, objectKey, file, header.Size, header.Header.Get("Content-Type")); err != nil {
\t\tlog.Printf("upload failed: %v", err)
\t\tc.JSON(http.StatusInternalServerError, gin.H{"error": "unable to store file"})
\t\treturn
\t}
\tu := models.Upload{
\t\tFilename:        filename,
\t\tOriginalName:    header.Filename,
\t\tStorageKey:      objectKey,
\t\tUploaderID:      userID,
\t\tStatus:          string(models.JobStatusReceived),
\t\tContentType:     header.Header.Get("Content-Type"),
\t\tSizeBytes:       header.Size,
\t\tHasPasswordHint: password != "",
\t}
\tuuidStr, err := api.repo.CreateUpload(context.Background(), &u)
\tif err != nil {
\t\tc.JSON(http.StatusInternalServerError, gin.H{"error": "db error creating upload"})
\t\treturn
\t}
\tjob := models.Job{
\t\tUploadID:     uuidStr,
\t\tStatus:       string(models.JobStatusQueued),
\t\tStage:        "queued",
\t\tAttemptCount: 0,
\t}
\tjobID, err := api.repo.CreateJob(context.Background(), &job)
\tif err != nil {
\t\tc.JSON(http.StatusInternalServerError, gin.H{"error": "db error creating job"})
\t\treturn
\t}
\tif err := api.repo.RecordAuditEvent(context.Background(), uuidStr, jobID, userID, "upload.created", map[string]any{"filename": filename}); err != nil {
\t\tlog.Printf("audit write failed: %v", err)
\t}
\tif err := api.repo.RecordAuditEvent(context.Background(), uuidStr, jobID, userID, "job.enqueued", map[string]any{"attempt": 0}); err != nil {
\t\tlog.Printf("audit write failed: %v", err)
\t}
\tif err := api.queue.Enqueue(context.Background(), models.JobPayload{
\t\tJobID:     jobID,
\t\tUploadID:  uuidStr,
\t\tObjectKey: objectKey,
\t\tFilename:  filename,
\t\tPassword:  password,
\t}); err != nil {
\t\tapi.repo.SetUploadAndJobStatus(context.Background(), uuidStr, jobID, string(models.JobStatusFailed), "failed to queue job")
\t\tc.JSON(http.StatusInternalServerError, gin.H{"error": "failed to queue job"})
\t\treturn
\t}
\tif err := api.repo.SetUploadStatus(context.Background(), uuidStr, string(models.JobStatusQueued), ""); err != nil {
\t\tlog.Printf("status update failed: %v", err)
\t}
\tc.JSON(http.StatusCreated, gin.H{
\t\t"upload_id": uuidStr,
\t\t"job_id":    jobID,
\t\t"status":    models.JobStatusQueued,
\t\t"filename":  filename,
\t})
}

func (api *API) listUploads(c *gin.Context) {
\tlimit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
\toffset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
\titems, err := api.repo.ListUploads(context.Background(), limit, offset)
\tif err != nil {
\t\tc.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list uploads"})
\t\treturn
\t}
\tc.JSON(http.StatusOK, items)
}

func (api *API) getUpload(c *gin.Context) {
\tid := c.Param("id")
\tu, err := api.repo.GetUpload(context.Background(), id)
\tif err != nil {
\t\tif err == pgx.ErrNoRows {
\t\t\tc.JSON(http.StatusNotFound, gin.H{"error": "upload not found"})
\t\t\treturn
\t\t}
\t\tc.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get upload"})
\t\treturn
\t}
\tfiles, err := api.repo.GetFilesByUpload(context.Background(), id)
\tif err != nil {
\t\tc.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get file tree"})
\t\treturn
\t}
\tvar job *models.Job
\tjob, _ = api.repo.LastJobForUpload(context.Background(), id)
\tc.JSON(http.StatusOK, gin.H{
\t\t"upload": u,
\t\t"job":    job,
\t\t"files":  files,
\t})
}

func (api *API) listJobs(c *gin.Context) {
\tlimit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
\toffset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
\titems, err := api.repo.ListJobs(context.Background(), limit, offset)
\tif err != nil {
\t\tc.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list jobs"})
\t\treturn
\t}
\tc.JSON(http.StatusOK, items)
}

func (api *API) getJob(c *gin.Context) {
\tjob, err := api.repo.GetJob(context.Background(), c.Param("id"))
\tif err != nil {
\t\tif err == pgx.ErrNoRows {
\t\t\tc.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
\t\t\treturn
\t\t}
\t\tc.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get job"})
\t\treturn
\t}
\tc.JSON(http.StatusOK, job)
}

func (api *API) retryWithPassword(c *gin.Context) {
\tjob, err := api.repo.GetJob(context.Background(), c.Param("id"))
\tif err != nil {
\t\tif err == pgx.ErrNoRows {
\t\t\tc.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
\t\t\treturn
\t\t}
\t\tc.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read job"})
\t\treturn
\t}
\tu, err := api.repo.GetUpload(context.Background(), job.UploadID)
\tif err != nil {
\t\tc.JSON(http.StatusNotFound, gin.H{"error": "upload not found"})
\t\treturn
\t}
\tvar req struct {
\t\tPassword string `json:"password"`
\t}
\tif err := c.ShouldBindJSON(&req); err != nil {
\t\tc.JSON(http.StatusBadRequest, gin.H{"error": "password is required"})
\t\treturn
\t}
\tif err := api.repo.SetJobStatus(context.Background(), job.ID, string(models.JobStatusQueued), "queued", ""); err != nil {
\t\tc.JSON(http.StatusInternalServerError, gin.H{"error": "could not update job"})
\t\treturn
\t}
\tif err := api.repo.SetUploadStatus(context.Background(), u.ID, string(models.JobStatusQueued), ""); err != nil {
\t\tc.JSON(http.StatusInternalServerError, gin.H{"error": "could not update upload"})
\t\treturn
\t}
\tif err := api.queue.Enqueue(context.Background(), models.JobPayload{
\t\tJobID:     job.ID,
\t\tUploadID:  job.UploadID,
\t\tObjectKey: u.StorageKey,
\t\tFilename:  u.Filename,
\t\tPassword:  req.Password,
\t\tAttempts:  job.AttemptCount + 1,
\t}); err != nil {
\t\tc.JSON(http.StatusInternalServerError, gin.H{"error": "could not enqueue retry"})
\t\treturn
\t}
\tif err := api.repo.RecordAuditEvent(context.Background(), u.ID, job.ID, middleware.UserID(c), "job.retry", map[string]any{"password_hint": req.Password != ""}); err != nil {
\t\tlog.Printf("audit fail: %v", err)
\t}
\tc.JSON(http.StatusAccepted, gin.H{"job_id": job.ID, "status": string(models.JobStatusQueued)})
}

func (api *API) searchContent(c *gin.Context) {
\tq := strings.TrimSpace(c.Query("q"))
\tif q == "" {
\t\tc.JSON(http.StatusBadRequest, gin.H{"error": "q is required"})
\t\treturn
\t}
\tif api.search == nil {
\t\tc.JSON(http.StatusServiceUnavailable, gin.H{"error": "search unavailable"})
\t\treturn
\t}
\tlimit, _ := strconv.Atoi(c.DefaultQuery("limit", "25"))
\tresults, err := api.search.Search(c.Request.Context(), q, limit)
\tif err != nil {
\t\tc.JSON(http.StatusInternalServerError, gin.H{"error": "search backend unavailable"})
\t\treturn
\t}
\tc.JSON(http.StatusOK, results)
}

func (api *API) fileTree(c *gin.Context) {
\tuuid := c.Param("uploadId")
\tfiles, err := api.repo.GetFilesByUpload(context.Background(), uuid)
\tif err != nil {
\t\tc.JSON(http.StatusInternalServerError, gin.H{"error": "unable to load tree"})
\t\treturn
\t}
\tindex := map[string]*models.FileRecord{}
\troots := make([]*models.FileRecord, 0)
\tfor i := range files {
\t\tf := files[i]
\t\tcopy := f
\t\tcopy.Children = make([]*models.FileRecord, 0)
\t\tindex[f.ID] = &copy
\t}
\tfor _, node := range index {
\t\tif node.ParentID == nil || *node.ParentID == "" {
\t\t\troots = append(roots, node)
\t\t\tcontinue
\t\t}
\t\tparent, ok := index[*node.ParentID]
\t\tif !ok {
\t\t\troots = append(roots, node)
\t\t\tcontinue
\t\t}
\t\tparent.Children = append(parent.Children, node)
\t}
\tc.JSON(http.StatusOK, roots)
}

func (api *API) dashboardSummary(c *gin.Context) {
\ts, err := api.repo.GetDashboardSummary(context.Background())
\tif err != nil {
\t\tc.JSON(http.StatusInternalServerError, gin.H{"error": "could not load summary"})
\t\treturn
\t}
\tc.JSON(http.StatusOK, s)
}

func (api *API) listAuditLogs(c *gin.Context) {
\tlimit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
\toffset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
\titems, err := api.repo.GetAuditLogs(context.Background(), limit, offset)
\tif err != nil {
\t\tc.JSON(http.StatusInternalServerError, gin.H{"error": "could not load audit logs"})
\t\treturn
\t}
\tc.JSON(http.StatusOK, items)
}

func (api *API) getSettings(c *gin.Context) {
\tapi.settingsMu.RLock()
\tdefer api.settingsMu.RUnlock()
\tc.JSON(http.StatusOK, api.settings)
}

func (api *API) updateSettings(c *gin.Context) {
\tvar req map[string]any
\tif err := c.ShouldBindJSON(&req); err != nil {
\t\tc.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
\t\treturn
\t}
\tapi.settingsMu.Lock()
\tdefer api.settingsMu.Unlock()
\tfor k, v := range req {
\t\tapi.settings[k] = v
\t}
\tc.JSON(http.StatusOK, api.settings)
}
