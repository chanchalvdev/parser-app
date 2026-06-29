package models

import "time"

type DashboardSummary struct {
	TenantID            string `json:"tenant_id"`
	TotalUploads        int64  `json:"total_uploads"`
	TotalFiles          int64  `json:"total_files"`
	TotalExtractedFiles int64  `json:"total_extracted_files"`
	TotalParsedRecords  int64  `json:"total_parsed_records"`
	CompletedJobs       int64  `json:"completed_jobs"`
	FailedJobs          int64  `json:"failed_jobs"`
	PasswordRequired    int64  `json:"password_required_files"`
	QuarantinedFiles    int64  `json:"quarantined_files"`
}

type DashboardBucket struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
}

type DashboardDistribution struct {
	TenantID string            `json:"tenant_id"`
	Buckets []*DashboardBucket `json:"buckets"`
}

type UploadVolumeBucket struct {
	Bucket       string `json:"bucket"`
	Uploads      int64  `json:"uploads"`
	Files        int64  `json:"files"`
	ParsedRecords int64 `json:"parsed_records"`
}

type DashboardUploadVolume struct {
	TenantID string              `json:"tenant_id"`
	Unit     string              `json:"unit"`
	Days     int                 `json:"days"`
	From     string              `json:"from"`
	To       string              `json:"to"`
	Buckets  []*UploadVolumeBucket `json:"buckets"`
}

type DashboardErrorBreakdownItem struct {
	ErrorCode string    `json:"error_code"`
	Count     int64     `json:"count"`
	LastSeen  time.Time `json:"last_seen"`
}

type DashboardErrorBreakdown struct {
	TenantID string                       `json:"tenant_id"`
	Total    int64                        `json:"total"`
	Errors   []*DashboardErrorBreakdownItem `json:"errors"`
}

type DashboardEntities struct {
	TenantID string                         `json:"tenant_id"`
	Limit    int                            `json:"limit"`
	Entities map[string][]*DashboardBucket   `json:"entities"`
}

type DashboardProcessingDuration struct {
	TenantID       string  `json:"tenant_id"`
	Unit           string  `json:"unit"`
	CompletedJobs  int64   `json:"completed_jobs"`
	AverageSeconds float64 `json:"average_seconds"`
	MedianSeconds  float64 `json:"median_seconds"`
	P95Seconds     float64 `json:"p95_seconds"`
	MinSeconds     float64 `json:"min_seconds"`
	MaxSeconds     float64 `json:"max_seconds"`
}
