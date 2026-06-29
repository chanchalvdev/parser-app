package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/enterprise-file-platform/api/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ParsedRecordRepository interface {
	ListRecordsByFileID(ctx context.Context, fileID string, limit int, offset int) ([]*models.ParsedRecord, error)
	CountRecordsByFileID(ctx context.Context, fileID string) (int64, error)
}

type PostgresParsedRecordRepository struct {
	pool *pgxpool.Pool
}

func NewParsedRecordRepository(pool *pgxpool.Pool) *PostgresParsedRecordRepository {
	return &PostgresParsedRecordRepository{pool: pool}
}

func (r *PostgresParsedRecordRepository) ListRecordsByFileID(ctx context.Context, fileID string, limit int, offset int) ([]*models.ParsedRecord, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if fileID == "" {
		return nil, fmt.Errorf("file_id is required")
	}
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT
			id, tenant_id, file_id, job_id, record_type, record_number, line_number, chunk_number,
			start_line, end_line, content_text, structured_data, extracted_entities, event_timestamp, created_at
		FROM parsed_records
		WHERE file_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, fileID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*models.ParsedRecord, 0)
	for rows.Next() {
		record, err := scanParsedRecord(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *PostgresParsedRecordRepository) CountRecordsByFileID(ctx context.Context, fileID string) (int64, error) {
	if err := validatePool(r.pool); err != nil {
		return 0, err
	}
	if fileID == "" {
		return 0, fmt.Errorf("file_id is required")
	}
	query := `SELECT COUNT(*) FROM parsed_records WHERE file_id = $1`
	var count int64
	err := r.pool.QueryRow(ctx, query, fileID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func scanParsedRecord(scanner rowScanner) (*models.ParsedRecord, error) {
	var (
		id, tenantID, fileID, jobID                             string
		recordType, contentText                                  sql.NullString
		recordNumber, lineNumber, chunkNumber, startLine, endLine sql.NullInt64
		structuredData, extractedEntities                        []byte
		eventTimestamp, createdAt                                sql.NullTime
	)

	if err := scanner.Scan(
		&id,
		&tenantID,
		&fileID,
		&jobID,
		&recordType,
		&recordNumber,
		&lineNumber,
		&chunkNumber,
		&startLine,
		&endLine,
		&contentText,
		&structuredData,
		&extractedEntities,
		&eventTimestamp,
		&createdAt,
	); err != nil {
		return nil, mapNoRows(err)
	}

	parsed, err := scanJSON(structuredData)
	if err != nil {
		return nil, err
	}
	extracted, err := scanJSON(extractedEntities)
	if err != nil {
		return nil, err
	}

	createdAtTime := createdAt.Time
	if !createdAt.Valid {
		createdAtTime = time.Time{}
	}
	var eventTime *time.Time
	if eventTimestamp.Valid {
		value := eventTimestamp.Time
		eventTime = &value
	}

	return &models.ParsedRecord{
		ID:                id,
		TenantID:          tenantID,
		FileID:            fileID,
		JobID:             jobID,
		RecordType:        nullPtrFromString(recordType),
		RecordNumber:      nullPtrFromInt64(recordNumber),
		LineNumber:        nullPtrFromInt64(lineNumber),
		ChunkNumber:       nullPtrFromInt64(chunkNumber),
		StartLine:         nullPtrFromInt64(startLine),
		EndLine:           nullPtrFromInt64(endLine),
		ContentText:       nullPtrFromString(contentText),
		StructuredData:    parsed,
		ExtractedEntities: extracted,
		EventTimestamp:    eventTime,
		CreatedAt:         createdAtTime,
	}, nil
}

func nullPtrFromInt64(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	value := v.Int64
	return &value
}
