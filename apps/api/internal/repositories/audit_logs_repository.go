package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/enterprise-file-platform/api/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditLogRepository interface {
	CreateAuditLog(ctx context.Context, log *models.AuditLog) (*models.AuditLog, error)
	ListAuditLogs(ctx context.Context, tenantID string, limit int, offset int) ([]*models.AuditLog, error)
}

type PostgresAuditLogRepository struct {
	pool *pgxpool.Pool
}

func NewAuditLogRepository(pool *pgxpool.Pool) *PostgresAuditLogRepository {
	return &PostgresAuditLogRepository{pool: pool}
}

func (r *PostgresAuditLogRepository) CreateAuditLog(ctx context.Context, logEntry *models.AuditLog) (*models.AuditLog, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if logEntry == nil {
		return nil, fmt.Errorf("audit log entry is required")
	}
	if logEntry.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if logEntry.Action == "" {
		return nil, fmt.Errorf("action is required")
	}
	if logEntry.ID == "" {
		logEntry.ID = uuid.NewString()
	}

	details, err := asJSON(logEntry.Details)
	if err != nil {
		return nil, fmt.Errorf("log details: %w", err)
	}
	ipAddress := nullString(logEntry.IPAddress)
	userAgent := nullString(logEntry.UserAgent)

	query := `
		INSERT INTO audit_logs (
			id, tenant_id, actor_user_id, action, entity_type, entity_id,
			details, ip_address, user_agent
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, tenant_id, actor_user_id, action, entity_type, entity_id, details, ip_address, user_agent, created_at
	`
	row := r.pool.QueryRow(
		ctx,
		query,
		logEntry.ID,
		logEntry.TenantID,
		nullString(logEntry.ActorUserID),
		logEntry.Action,
		nullString(logEntry.EntityType),
		nullString(logEntry.EntityID),
		details,
		ipAddress,
		userAgent,
	)
	return scanAuditLog(row)
}

func (r *PostgresAuditLogRepository) ListAuditLogs(ctx context.Context, tenantID string, limit int, offset int) ([]*models.AuditLog, error) {
	if err := validatePool(r.pool); err != nil {
		return nil, err
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, tenant_id, actor_user_id, action, entity_type, entity_id, details, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*models.AuditLog, 0)
	for rows.Next() {
		entry, err := scanAuditLog(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func scanAuditLog(scanner rowScanner) (*models.AuditLog, error) {
	var (
		id, tenantID, action                               string
		actorUserID, entityType, entityID, ipAddress, userAgent sql.NullString
		details                                            []byte
		createdAt                                          sql.NullTime
	)
	if err := scanner.Scan(
		&id,
		&tenantID,
		&actorUserID,
		&action,
		&entityType,
		&entityID,
		&details,
		&ipAddress,
		&userAgent,
		&createdAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	parsedDetails, err := scanJSON(details)
	if err != nil {
		return nil, err
	}
	createdAtTime := createdAt.Time
	if !createdAt.Valid {
		createdAtTime = time.Time{}
	}

	return &models.AuditLog{
		ID:          id,
		TenantID:    tenantID,
		ActorUserID: nullPtrFromString(actorUserID),
		Action:      action,
		EntityType:  nullPtrFromString(entityType),
		EntityID:    nullPtrFromString(entityID),
		Details:     parsedDetails,
		IPAddress:   nullPtrFromString(ipAddress),
		UserAgent:   nullPtrFromString(userAgent),
		CreatedAt:   createdAtTime,
	}, nil
}

