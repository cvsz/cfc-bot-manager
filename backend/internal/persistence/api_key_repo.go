package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/cvsz/cfc-bot-manager/backend/internal/domain"
	"github.com/cvsz/cfc-bot-manager/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type apiKeyRepo struct {
	pool *pgxpool.Pool
}

func NewAPIKeyRepository(pool *pgxpool.Pool) *apiKeyRepo {
	return &apiKeyRepo{pool: pool}
}

func (r *apiKeyRepo) Create(ctx context.Context, key *domain.APIKey) error {
	query := `
		INSERT INTO api_keys (id, organization_id, name, key_hash, prefix, permissions, expires_at, last_used_at, revoked_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.pool.Exec(ctx, query,
		key.ID, key.OrganizationID, key.Name, key.KeyHash,
		key.Prefix, key.Permissions, key.ExpiresAt,
		key.LastUsedAt, key.RevokedAt, key.CreatedAt,
	)
	return err
}

func (r *apiKeyRepo) GetByID(ctx context.Context, id domain.APIKeyID) (*domain.APIKey, error) {
	query := `
		SELECT id, organization_id, name, key_hash, prefix, permissions, expires_at, last_used_at, revoked_at, created_at
		FROM api_keys WHERE id = $1
	`
	var key domain.APIKey
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&key.ID, &key.OrganizationID, &key.Name, &key.KeyHash,
		&key.Prefix, &key.Permissions, &key.ExpiresAt,
		&key.LastUsedAt, &key.RevokedAt, &key.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrAPIKeyNotFound
	}
	return &key, err
}

func (r *apiKeyRepo) GetByPrefix(ctx context.Context, orgID domain.OrganizationID, prefix string) (*domain.APIKey, error) {
	query := `
		SELECT id, organization_id, name, key_hash, prefix, permissions, expires_at, last_used_at, revoked_at, created_at
		FROM api_keys WHERE organization_id = $1 AND prefix = $2
	`
	var key domain.APIKey
	err := r.pool.QueryRow(ctx, query, orgID, prefix).Scan(
		&key.ID, &key.OrganizationID, &key.Name, &key.KeyHash,
		&key.Prefix, &key.Permissions, &key.ExpiresAt,
		&key.LastUsedAt, &key.RevokedAt, &key.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrAPIKeyNotFound
	}
	return &key, err
}

func (r *apiKeyRepo) ListByOrg(ctx context.Context, orgID domain.OrganizationID, pagination repository.Pagination) (repository.ListResult[domain.APIKey], error) {
	countQuery := `SELECT COUNT(*) FROM api_keys WHERE organization_id = $1`
	var totalCount int64
	err := r.pool.QueryRow(ctx, countQuery, orgID).Scan(&totalCount)
	if err != nil {
		return repository.ListResult[domain.APIKey]{}, err
	}

	query := `
		SELECT id, organization_id, name, key_hash, prefix, permissions, expires_at, last_used_at, revoked_at, created_at
		FROM api_keys WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, orgID, pagination.Limit, pagination.Offset)
	if err != nil {
		return repository.ListResult[domain.APIKey]{}, err
	}
	defer rows.Close()

	var keys []domain.APIKey
	for rows.Next() {
		var key domain.APIKey
		if err := rows.Scan(
			&key.ID, &key.OrganizationID, &key.Name, &key.KeyHash,
			&key.Prefix, &key.Permissions, &key.ExpiresAt,
			&key.LastUsedAt, &key.RevokedAt, &key.CreatedAt,
		); err != nil {
			return repository.ListResult[domain.APIKey]{}, err
		}
		keys = append(keys, key)
	}

	return repository.ListResult[domain.APIKey]{
		Items:      keys,
		TotalCount: totalCount,
		Pagination: pagination,
	}, nil
}

func (r *apiKeyRepo) Update(ctx context.Context, key *domain.APIKey) error {
	query := `
		UPDATE api_keys
		SET name = $2, permissions = $3, expires_at = $4, revoked_at = $5
		WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query,
		key.ID, key.Name, key.Permissions, key.ExpiresAt, key.RevokedAt,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrAPIKeyNotFound
	}
	return nil
}

func (r *apiKeyRepo) UpdateLastUsed(ctx context.Context, id domain.APIKeyID, t time.Time) error {
	query := `UPDATE api_keys SET last_used_at = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, t, id)
	return err
}

func (r *apiKeyRepo) Revoke(ctx context.Context, id domain.APIKeyID) error {
	query := `UPDATE api_keys SET revoked_at = $1 WHERE id = $2`
	result, err := r.pool.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrAPIKeyNotFound
	}
	return nil
}

func (r *apiKeyRepo) Delete(ctx context.Context, id domain.APIKeyID) error {
	query := `DELETE FROM api_keys WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrAPIKeyNotFound
	}
	return nil
}

var _ repository.APIKeyRepository = (*apiKeyRepo)(nil)