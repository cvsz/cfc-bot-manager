package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/cvsz/cfc-bot-manager/backend/internal/domain"
	"github.com/cvsz/cfc-bot-manager/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type sessionRepo struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *sessionRepo {
	return &sessionRepo{pool: pool}
}

func (r *sessionRepo) Create(ctx context.Context, session *domain.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, organization_id, token_hash, csrf_token, expires_at, revoked_at, user_agent, ip_address, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.pool.Exec(ctx, query,
		session.ID, session.UserID, session.Organization, session.TokenHash,
		session.CSRFToken, session.ExpiresAt, session.RevokedAt,
		session.UserAgent, session.IPAddress, session.CreatedAt,
	)
	return err
}

func (r *sessionRepo) GetByID(ctx context.Context, id domain.SessionID) (*domain.Session, error) {
	query := `
		SELECT id, user_id, organization_id, token_hash, csrf_token, expires_at, revoked_at, user_agent, ip_address, created_at
		FROM sessions WHERE id = $1
	`
	var s domain.Session
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.UserID, &s.Organization, &s.TokenHash, &s.CSRFToken,
		&s.ExpiresAt, &s.RevokedAt, &s.UserAgent, &s.IPAddress, &s.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrSessionNotFound
	}
	return &s, err
}

func (r *sessionRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	query := `
		SELECT id, user_id, organization_id, token_hash, csrf_token, expires_at, revoked_at, user_agent, ip_address, created_at
		FROM sessions WHERE token_hash = $1
	`
	var s domain.Session
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&s.ID, &s.UserID, &s.Organization, &s.TokenHash, &s.CSRFToken,
		&s.ExpiresAt, &s.RevokedAt, &s.UserAgent, &s.IPAddress, &s.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrSessionNotFound
	}
	return &s, err
}

func (r *sessionRepo) Update(ctx context.Context, session *domain.Session) error {
	query := `
		UPDATE sessions
		SET token_hash = $2, csrf_token = $3, expires_at = $4, revoked_at = $5, user_agent = $6, ip_address = $7
		WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query,
		session.ID, session.TokenHash, session.CSRFToken,
		session.ExpiresAt, session.RevokedAt, session.UserAgent, session.IPAddress,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrSessionNotFound
	}
	return nil
}

func (r *sessionRepo) Revoke(ctx context.Context, id domain.SessionID) error {
	query := `UPDATE sessions SET revoked_at = $1 WHERE id = $2`
	result, err := r.pool.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrSessionNotFound
	}
	return nil
}

func (r *sessionRepo) RevokeAllForUser(ctx context.Context, userID domain.UserID) error {
	query := `UPDATE sessions SET revoked_at = $1 WHERE user_id = $2 AND revoked_at IS NULL`
	_, err := r.pool.Exec(ctx, query, time.Now(), userID)
	return err
}

func (r *sessionRepo) RevokeAllForOrg(ctx context.Context, orgID domain.OrganizationID) error {
	query := `UPDATE sessions SET revoked_at = $1 WHERE organization_id = $2 AND revoked_at IS NULL`
	_, err := r.pool.Exec(ctx, query, time.Now(), orgID)
	return err
}

func (r *sessionRepo) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM sessions WHERE expires_at < $1 OR revoked_at IS NOT NULL`
	result, err := r.pool.Exec(ctx, query, time.Now())
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (r *sessionRepo) CleanupRevoked(ctx context.Context, olderThan time.Time) (int64, error) {
	query := `DELETE FROM sessions WHERE revoked_at IS NOT NULL AND revoked_at < $1`
	result, err := r.pool.Exec(ctx, query, olderThan)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

var _ repository.SessionRepository = (*sessionRepo)(nil)