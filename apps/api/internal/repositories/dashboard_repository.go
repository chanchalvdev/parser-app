package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/enterprise-file-platform/api/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DashboardRepository interface {
	GetDashboardSummary(ctx context.Context, tenantID string) (*models.DashboardSummary, error)
	GetDashboardFileTypeDistribution(ctx context.Context, tenantID string, limit int) ([]*models.DashboardBucket, error)
	GetDashboardProcessingStatusDistribution(ctx context.Context, tenantID string, limit int) ([]*models.DashboardBucket, error)
	GetDashboardUploadVolume(ctx context.Context, tenantID string, from time.Time, to time.Time) ([]*models.UploadVolumeBucket, error)
	GetDashboardErrorBreakdown(ctx context.Context, tenantID string, limit int) (*models.DashboardErrorBreakdown, error)
	GetDashboardTopEntities(ctx context.Context, tenantID string, limit int) (map[string][]*models.DashboardBucket, error)
	GetDashboardProcessingDuration(ctx context.Context, tenantID string) (*models.DashboardProcessingDuration, error)
}

type PostgresDashboardRepository struct {
	pool *pgxpool.Pool
}

func NewDashboardRepository(pool *pgxpool.Pool) *PostgresDashboardRepository {
	return &PostgresDashboardRepository{pool: pool}
}

func (r *PostgresDashboardRepository) GetDashboardSummary(ctx context.Context, tenantID string) (*models.DashboardSummary, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	row := r.pool.QueryRow(
		ctx,
		`
		SELECT
			COALESCE((SELECT COUNT(*) FROM uploads WHERE tenant_id = $1), 0),
			COALESCE((SELECT COUNT(*) FROM files WHERE tenant_id = $1), 0),
			COALESCE((SELECT COUNT(*) FROM files WHERE tenant_id = $1 AND processing_status = 'completed'), 0),
			COALESCE((SELECT COUNT(*) FROM parsed_records WHERE tenant_id = $1), 0),
			COALESCE((SELECT COUNT(*) FROM ingestion_jobs WHERE tenant_id = $1 AND status = 'completed'), 0),
			COALESCE((SELECT COUNT(*) FROM ingestion_jobs WHERE tenant_id = $1 AND status = 'failed'), 0),
			COALESCE((SELECT COUNT(*) FROM files WHERE tenant_id = $1 AND processing_status = 'password_required'), 0),
			COALESCE((SELECT COUNT(*) FROM files WHERE tenant_id = $1 AND processing_status = 'quarantined'), 0)
		`,
		tenantID,
	)

	var summary models.DashboardSummary
	summary.TenantID = tenantID
	if err := row.Scan(
		&summary.TotalUploads,
		&summary.TotalFiles,
		&summary.TotalExtractedFiles,
		&summary.TotalParsedRecords,
		&summary.CompletedJobs,
		&summary.FailedJobs,
		&summary.PasswordRequired,
		&summary.QuarantinedFiles,
	); err != nil {
		return nil, err
	}

	return &summary, nil
}

func (r *PostgresDashboardRepository) GetDashboardFileTypeDistribution(ctx context.Context, tenantID string, limit int) ([]*models.DashboardBucket, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT
			COALESCE(NULLIF(TRIM(detected_file_type), ''), 'unknown') AS value,
			COUNT(*) AS count
		FROM files
		WHERE tenant_id = $1
		GROUP BY COALESCE(NULLIF(TRIM(detected_file_type), ''), 'unknown')
		ORDER BY COUNT(*) DESC, value ASC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, query, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*models.DashboardBucket, 0)
	for rows.Next() {
		bucket := &models.DashboardBucket{}
		if err := rows.Scan(&bucket.Value, &bucket.Count); err != nil {
			return nil, err
		}
		result = append(result, bucket)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *PostgresDashboardRepository) GetDashboardProcessingStatusDistribution(ctx context.Context, tenantID string, limit int) ([]*models.DashboardBucket, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT
			COALESCE(NULLIF(TRIM(processing_status), ''), 'unknown') AS value,
			COUNT(*) AS count
		FROM files
		WHERE tenant_id = $1
		GROUP BY COALESCE(NULLIF(TRIM(processing_status), ''), 'unknown')
		ORDER BY COUNT(*) DESC, value ASC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, query, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*models.DashboardBucket, 0)
	for rows.Next() {
		bucket := &models.DashboardBucket{}
		if err := rows.Scan(&bucket.Value, &bucket.Count); err != nil {
			return nil, err
		}
		result = append(result, bucket)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *PostgresDashboardRepository) GetDashboardUploadVolume(ctx context.Context, tenantID string, from time.Time, to time.Time) ([]*models.UploadVolumeBucket, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	query := `
		WITH date_series AS (
			SELECT generate_series($1::date, $2::date, interval '1 day') AS bucket
		),
		upload_buckets AS (
			SELECT created_at::date AS bucket_day, COUNT(*) AS count
			FROM uploads
			WHERE tenant_id = $3 AND created_at::date BETWEEN $1::date AND $2::date
			GROUP BY created_at::date
		),
		file_buckets AS (
			SELECT created_at::date AS bucket_day, COUNT(*) AS count
			FROM files
			WHERE tenant_id = $3 AND created_at::date BETWEEN $1::date AND $2::date
			GROUP BY created_at::date
		),
		record_buckets AS (
			SELECT created_at::date AS bucket_day, COUNT(*) AS count
			FROM parsed_records
			WHERE tenant_id = $3 AND created_at::date BETWEEN $1::date AND $2::date
			GROUP BY created_at::date
		)
		SELECT
			ds.bucket::date AS bucket,
			COALESCE(u.count, 0) AS uploads,
			COALESCE(f.count, 0) AS files,
			COALESCE(p.count, 0) AS parsed_records
		FROM date_series ds
		LEFT JOIN upload_buckets u ON ds.bucket::date = u.bucket_day
		LEFT JOIN file_buckets f ON ds.bucket::date = f.bucket_day
		LEFT JOIN record_buckets p ON ds.bucket::date = p.bucket_day
		ORDER BY ds.bucket
	`
	rows, err := r.pool.Query(ctx, query, from, to, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*models.UploadVolumeBucket, 0)
	for rows.Next() {
		item := &models.UploadVolumeBucket{}
		var bucket time.Time
		if err := rows.Scan(&bucket, &item.Uploads, &item.Files, &item.ParsedRecords); err != nil {
			return nil, err
		}
		item.Bucket = bucket.Format("2006-01-02")
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *PostgresDashboardRepository) GetDashboardErrorBreakdown(ctx context.Context, tenantID string, limit int) (*models.DashboardErrorBreakdown, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if limit <= 0 {
		limit = 10
	}

	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM parser_errors WHERE tenant_id = $1`, tenantID).Scan(&total); err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(
		ctx,
		`
		SELECT
			COALESCE(NULLIF(error_code, ''), 'unknown') AS error_code,
			COUNT(*) AS count,
			MAX(occurred_at) AS last_seen
		FROM parser_errors
		WHERE tenant_id = $1
		GROUP BY COALESCE(NULLIF(error_code, ''), 'unknown')
		ORDER BY COUNT(*) DESC, error_code ASC
		LIMIT $2
		`,
		tenantID,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	errorsByType := make([]*models.DashboardErrorBreakdownItem, 0)
	for rows.Next() {
		item := &models.DashboardErrorBreakdownItem{}
		var lastSeen sql.NullTime
		if err := rows.Scan(&item.ErrorCode, &item.Count, &lastSeen); err != nil {
			return nil, err
		}
		if lastSeen.Valid {
			item.LastSeen = lastSeen.Time
		}
		errorsByType = append(errorsByType, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &models.DashboardErrorBreakdown{
		TenantID: tenantID,
		Total:    total,
		Errors:   errorsByType,
	}, nil
}

func (r *PostgresDashboardRepository) GetDashboardTopEntities(ctx context.Context, tenantID string, limit int) (map[string][]*models.DashboardBucket, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if limit <= 0 {
		limit = 10
	}

	query := `
		WITH extracted AS (
			SELECT 'ip_addresses' AS entity_type, jsonb_array_elements_text(
				CASE
					WHEN jsonb_typeof(extracted_entities->'ip_addresses') = 'array' THEN extracted_entities->'ip_addresses'
					ELSE '[]'::jsonb
				END
			) AS value
			FROM parsed_records
			WHERE tenant_id = $1
			UNION ALL
			SELECT 'emails' AS entity_type, jsonb_array_elements_text(
				CASE
					WHEN jsonb_typeof(extracted_entities->'emails') = 'array' THEN extracted_entities->'emails'
					ELSE '[]'::jsonb
				END
			) AS value
			FROM parsed_records
			WHERE tenant_id = $1
			UNION ALL
			SELECT 'urls' AS entity_type, jsonb_array_elements_text(
				CASE
					WHEN jsonb_typeof(extracted_entities->'urls') = 'array' THEN extracted_entities->'urls'
					ELSE '[]'::jsonb
				END
			) AS value
			FROM parsed_records
			WHERE tenant_id = $1
			UNION ALL
			SELECT 'domains' AS entity_type, jsonb_array_elements_text(
				CASE
					WHEN jsonb_typeof(extracted_entities->'domains') = 'array' THEN extracted_entities->'domains'
					ELSE '[]'::jsonb
				END
			) AS value
			FROM parsed_records
			WHERE tenant_id = $1
			UNION ALL
			SELECT 'hashes' AS entity_type, jsonb_array_elements_text(
				CASE
					WHEN jsonb_typeof(extracted_entities->'hashes') = 'array' THEN extracted_entities->'hashes'
					ELSE '[]'::jsonb
				END
			) AS value
			FROM parsed_records
			WHERE tenant_id = $1
		),
		counted AS (
			SELECT
				entity_type,
				LOWER(TRIM(value)) AS value,
				COUNT(*) AS count
			FROM extracted
			WHERE TRIM(value) <> ''
			GROUP BY entity_type, LOWER(TRIM(value))
		),
		ranked AS (
			SELECT
				entity_type,
				value,
				count,
				ROW_NUMBER() OVER (PARTITION BY entity_type ORDER BY count DESC, value ASC) AS rn
			FROM counted
		)
		SELECT entity_type, value, count
		FROM ranked
		WHERE rn <= $2
		ORDER BY entity_type ASC, count DESC, value ASC
	`
	rows, err := r.pool.Query(ctx, query, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[string][]*models.DashboardBucket{
		"ip_addresses": make([]*models.DashboardBucket, 0),
		"emails":       make([]*models.DashboardBucket, 0),
		"urls":         make([]*models.DashboardBucket, 0),
		"domains":      make([]*models.DashboardBucket, 0),
		"hashes":       make([]*models.DashboardBucket, 0),
	}
	for rows.Next() {
		var entityType string
		bucket := &models.DashboardBucket{}
		if err := rows.Scan(&entityType, &bucket.Value, &bucket.Count); err != nil {
			return nil, err
		}
		result[entityType] = append(result[entityType], bucket)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *PostgresDashboardRepository) GetDashboardProcessingDuration(ctx context.Context, tenantID string) (*models.DashboardProcessingDuration, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	var (
		completedJobs int64
		avg           sql.NullFloat64
		median        sql.NullFloat64
		p95           sql.NullFloat64
		minDuration   sql.NullFloat64
		maxDuration   sql.NullFloat64
	)

	row := r.pool.QueryRow(
		ctx,
		`
		SELECT
			COUNT(*) FILTER (WHERE started_at IS NOT NULL AND completed_at IS NOT NULL),
			AVG(EXTRACT(EPOCH FROM (completed_at - started_at))) AS avg_seconds,
			PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (completed_at - started_at))) AS median_seconds,
			PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (completed_at - started_at))) AS p95_seconds,
			MIN(EXTRACT(EPOCH FROM (completed_at - started_at))) AS min_seconds,
			MAX(EXTRACT(EPOCH FROM (completed_at - started_at))) AS max_seconds
		FROM ingestion_jobs
		WHERE tenant_id = $1 AND status = 'completed' AND started_at IS NOT NULL AND completed_at IS NOT NULL
		`,
		tenantID,
	)
	if err := row.Scan(&completedJobs, &avg, &median, &p95, &minDuration, &maxDuration); err != nil {
		return nil, err
	}

	duration := &models.DashboardProcessingDuration{
		TenantID: tenantID,
		Unit:     "seconds",
	}
	duration.CompletedJobs = completedJobs
	if avg.Valid {
		duration.AverageSeconds = avg.Float64
	}
	if median.Valid {
		duration.MedianSeconds = median.Float64
	}
	if p95.Valid {
		duration.P95Seconds = p95.Float64
	}
	if minDuration.Valid {
		duration.MinSeconds = minDuration.Float64
	}
	if maxDuration.Valid {
		duration.MaxSeconds = maxDuration.Float64
	}

	return duration, nil
}
