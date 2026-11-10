package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/cvsz/cfc-bot-manager/backend/internal/domain"
	"github.com/cvsz/cfc-bot-manager/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *userRepo {
	return &userRepo{pool: pool}
}

func (r *userRepo) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, organization_id, email, display_name, password_hash, mfa_secret, mfa_enabled, is_active, last_login_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.pool.Exec(ctx, query,
		user.ID, user.Organization, user.Email, user.DisplayName, user.PasswordHash,
		user.MFASecret, user.MFAEnabled, user.IsActive, user.LastLoginAt,
		user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *userRepo) GetByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	query := `
		SELECT id, organization_id, email, display_name, password_hash, mfa_secret, mfa_enabled, is_active, last_login_at, created_at, updated_at
		FROM users WHERE id = $1
	`
	var user domain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Organization, &user.Email, &user.DisplayName, &user.PasswordHash,
		&user.MFASecret, &user.MFAEnabled, &user.IsActive, &user.LastLoginAt,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &user, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, orgID domain.OrganizationID, email string) (*domain.User, error) {
	query := `
		SELECT id, organization_id, email, display_name, password_hash, mfa_secret, mfa_enabled, is_active, last_login_at, created_at, updated_at
		FROM users WHERE organization_id = $1 AND email = $2
	`
	var user domain.User
	err := r.pool.QueryRow(ctx, query, orgID, email).Scan(
		&user.ID, &user.Organization, &user.Email, &user.DisplayName, &user.PasswordHash,
		&user.MFASecret, &user.MFAEnabled, &user.IsActive, &user.LastLoginAt,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &user, nil
}

func (r *userRepo) List(ctx context.Context, orgID domain.OrganizationID, pagination repository.Pagination, sort repository.Sort, filter repository.Filter) (repository.ListResult[domain.User], error) {
	whereClause, args := buildFilterClauseWithOrg(filter, orgID, 1)
	argIndex := len(args) + 1

	sortField := "created_at"
	if sort.Field != "" {
		allowedSortFields := map[string]bool{"email": true, "display_name": true, "created_at": true, "updated_at": true, "is_active": true}
		if allowedSortFields[sort.Field] {
			sortField = sort.Field
		}
	}
	sortOrder := "DESC"
	if sort.Order == repository.SortAsc {
		sortOrder = "ASC"
	}

	countQuery := "SELECT COUNT(*) FROM users" + whereClause
	var totalCount int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return repository.ListResult[domain.User]{}, fmt.Errorf("count users: %w", err)
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, organization_id, email, display_name, password_hash, mfa_secret, mfa_enabled, is_active, last_login_at, created_at, updated_at
		FROM users
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortField, sortOrder, argIndex, argIndex+1)

	args = append(args, pagination.Limit, pagination.Offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return repository.ListResult[domain.User]{}, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.ID, &user.Organization, &user.Email, &user.DisplayName, &user.PasswordHash,
			&user.MFASecret, &user.MFAEnabled, &user.IsActive, &user.LastLoginAt,
			&user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return repository.ListResult[domain.User]{}, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}

	return repository.ListResult[domain.User]{
		Items:      users,
		TotalCount: totalCount,
		Pagination: pagination,
	}, nil
}

func (r *userRepo) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET email = $2, display_name = $3, password_hash = $4, mfa_secret = $5, mfa_enabled = $6, is_active = $7, updated_at = $8
		WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query,
		user.ID, user.Email, user.DisplayName, user.PasswordHash,
		user.MFASecret, user.MFAEnabled, user.IsActive, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *userRepo) UpdateLastLogin(ctx context.Context, id domain.UserID, t time.Time) error {
	query := `UPDATE users SET last_login_at = $1, updated_at = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, t, time.Now(), id)
	return err
}

func (r *userRepo) Delete(ctx context.Context, id domain.UserID) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *userRepo) ExistsByEmail(ctx context.Context, orgID domain.OrganizationID, email string, excludeID *domain.UserID) (bool, error) {
	query := `SELECT 1 FROM users WHERE organization_id = $1 AND email = $2`
	args := []any{orgID, email}
	if excludeID != nil {
		query += ` AND id != $3`
		args = append(args, *excludeID)
	}
	var exists int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check email exists: %w", err)
	}
	return true, nil
}

func buildFilterClauseWithOrg(filter repository.Filter, orgID domain.OrganizationID, startIndex int) (string, []any) {
	conditions := []string{"organization_id = $1"}
	args := []any{orgID}
	argIndex := startIndex + 1

	for key, value := range filter {
		conditions = append(conditions, fmt.Sprintf("%s = $%d", key, argIndex))
		args = append(args, value)
		argIndex++
	}

	return " WHERE " + joinConditions(conditions), args
}

var _ repository.UserRepository = (*userRepo)(nil)