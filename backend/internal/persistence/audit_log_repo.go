package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/cvsz/cfc-bot-manager/backend/internal/domain"
	"github.com/cvsz/cfc-bot-manager/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type auditLogRepo struct {
	pool *pgxpool.Pool
}

func NewAuditLogRepository(pool *pgxpool.Pool) *auditLogRepo {
	return &auditLogRepo{pool: pool}
}

func (r *auditLogRepo) Create(ctx context.Context, log *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, organization_id, actor_user_id, actor_type, action, resource_type, resource_id, old_value, new_value, ip_address, user_agent, correlation_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.pool.Exec(ctx, query,
		log.ID, log.OrganizationID, log.ActorUserID, log.ActorType,
		log.Action, log.ResourceType, log.ResourceID,
		log.OldValue, log.NewValue, log.IPAddress, log.UserAgent,
		log.CorrelationID, log.CreatedAt,
	)
	return err
}

func (r *auditLogRepo) ListByOrg(ctx context.Context, orgID domain.OrganizationID, pagination repository.Pagination, sort repository.Sort, filter repository.Filter) (repository.ListResult[domain.AuditLog], error) {
	whereClause, args := buildFilterClauseWithOrg(filter, orgID, 1)
	argIndex := len(args) + 1

	sortField := "created_at"
	if sort.Field != "" {
		allowedSortFields := map[string]bool{"created_at": true, "action": true, "resource_type": true}
		if allowedSortFields[sort.Field] {
			sortField = sort.Field
		}
	}
	sortOrder := "DESC"
	if sort.Order == repository.SortAsc {
		sortOrder = "ASC"
	}

	countQuery := "SELECT COUNT(*) FROM audit_logs" + whereClause
	var totalCount int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return repository.ListResult[domain.AuditLog]{}, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, organization_id, actor_user_id, actor_type, action, resource_type, resource_id, old_value, new_value, ip_address, user_agent, correlation_id, created_at
		FROM audit_logs
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortField, sortOrder, argIndex, argIndex+1)

	args = append(args, pagination.Limit, pagination.Offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return repository.ListResult[domain.AuditLog]{}, err
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var log domain.AuditLog
		var actorUserID sql.NullString
		if err := rows.Scan(
			&log.ID, &log.OrganizationID, &actorUserID, &log.ActorType,
			&log.Action, &log.ResourceType, &log.ResourceID,
			&log.OldValue, &log.NewValue, &log.IPAddress, &log.UserAgent,
			&log.CorrelationID, &log.CreatedAt,
		); err != nil {
			return repository.ListResult[domain.AuditLog]{}, err
		}
		if actorUserID.Valid {
			uid := domain.UserID(uuid.MustParse(actorUserID.String))
			log.ActorUserID = &uid
		}
		logs = append(logs, log)
	}

	return repository.ListResult[domain.AuditLog]{
		Items:      logs,
		TotalCount: totalCount,
		Pagination: pagination,
	}, nil
}

func (r *auditLogRepo) ListByResource(ctx context.Context, orgID domain.OrganizationID, resourceType, resourceID string, pagination repository.Pagination) (repository.ListResult[domain.AuditLog], error) {
	query := `
		SELECT id, organization_id, actor_user_id, actor_type, action, resource_type, resource_id, old_value, new_value, ip_address, user_agent, correlation_id, created_at
		FROM audit_logs
		WHERE organization_id = $1 AND resource_type = $2 AND resource_id = $3
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`
	rows, err := r.pool.Query(ctx, query, orgID, resourceType, resourceID, pagination.Limit, pagination.Offset)
	if err != nil {
		return repository.ListResult[domain.AuditLog]{}, err
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var log domain.AuditLog
		var actorUserID sql.NullString
		if err := rows.Scan(
			&log.ID, &log.OrganizationID, &actorUserID, &log.ActorType,
			&log.Action, &log.ResourceType, &log.ResourceID,
			&log.OldValue, &log.NewValue, &log.IPAddress, &log.UserAgent,
			&log.CorrelationID, &log.CreatedAt,
		); err != nil {
			return repository.ListResult[domain.AuditLog]{}, err
		}
		if actorUserID.Valid {
			uid := domain.UserID(uuid.MustParse(actorUserID.String))
			log.ActorUserID = &uid
		}
		logs = append(logs, log)
	}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM audit_logs WHERE organization_id = $1 AND resource_type = $2 AND resource_id = $3`
	var totalCount int64
	err = r.pool.QueryRow(ctx, countQuery, orgID, resourceType, resourceID).Scan(&totalCount)
	if err != nil {
		return repository.ListResult[domain.AuditLog]{}, err
	}

	return repository.ListResult[domain.AuditLog]{
		Items:      logs,
		TotalCount: totalCount,
		Pagination: pagination,
	}, nil
}

func (r *auditLogRepo) ListByActor(ctx context.Context, orgID domain.OrganizationID, actorUserID domain.UserID, pagination repository.Pagination) (repository.ListResult[domain.AuditLog], error) {
	query := `
		SELECT id, organization_id, actor_user_id, actor_type, action, resource_type, resource_id, old_value, new_value, ip_address, user_agent, correlation_id, created_at
		FROM audit_logs
		WHERE organization_id = $1 AND actor_user_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.pool.Query(ctx, query, orgID, actorUserID, pagination.Limit, pagination.Offset)
	if err != nil {
		return repository.ListResult[domain.AuditLog]{}, err
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var log domain.AuditLog
		var aUID sql.NullString
		if err := rows.Scan(
			&log.ID, &log.OrganizationID, &aUID, &log.ActorType,
			&log.Action, &log.ResourceType, &log.ResourceID,
			&log.OldValue, &log.NewValue, &log.IPAddress, &log.UserAgent,
			&log.CorrelationID, &log.CreatedAt,
		); err != nil {
			return repository.ListResult[domain.AuditLog]{}, err
		}
		if aUID.Valid {
			uid := domain.UserID(uuid.MustParse(aUID.String))
			log.ActorUserID = &uid
		}
		logs = append(logs, log)
	}

	countQuery := `SELECT COUNT(*) FROM audit_logs WHERE organization_id = $1 AND actor_user_id = $2`
	var totalCount int64
	err = r.pool.QueryRow(ctx, countQuery, orgID, actorUserID).Scan(&totalCount)
	if err != nil {
		return repository.ListResult[domain.AuditLog]{}, err
	}

	return repository.ListResult[domain.AuditLog]{
		Items:      logs,
		TotalCount: totalCount,
		Pagination: pagination,
	}, nil
}

func (r *auditLogRepo) DeleteOlderThan(ctx context.Context, orgID domain.OrganizationID, olderThan time.Time) (int64, error) {
	query := `DELETE FROM audit_logs WHERE organization_id = $1 AND created_at < $2`
	result, err := r.pool.Exec(ctx, query, orgID, olderThan)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

var _ repository.AuditLogRepository = (*auditLogRepo)(nil)