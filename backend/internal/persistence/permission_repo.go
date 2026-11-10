package persistence

import (
	"context"
	"database/sql"

	"github.com/cvsz/cfc-bot-manager/backend/internal/domain"
	"github.com/cvsz/cfc-bot-manager/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type permissionRepo struct {
	pool *pgxpool.Pool
}

func NewPermissionRepository(pool *pgxpool.Pool) *permissionRepo {
	return &permissionRepo{pool: pool}
}

func (r *permissionRepo) Create(ctx context.Context, perm *domain.Permission) error {
	query := `INSERT INTO permissions (id, name, description, resource, action, created_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.pool.Exec(ctx, query, perm.ID, perm.Name, perm.Description, perm.Resource, perm.Action, perm.CreatedAt)
	return err
}

func (r *permissionRepo) GetByID(ctx context.Context, id domain.PermissionID) (*domain.Permission, error) {
	query := `SELECT id, name, description, resource, action, created_at FROM permissions WHERE id = $1`
	var p domain.Permission
	err := r.pool.QueryRow(ctx, query, id).Scan(&p.ID, &p.Name, &p.Description, &p.Resource, &p.Action, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrPermissionNotFound
	}
	return &p, err
}

func (r *permissionRepo) GetByName(ctx context.Context, name string) (*domain.Permission, error) {
	query := `SELECT id, name, description, resource, action, created_at FROM permissions WHERE name = $1`
	var p domain.Permission
	err := r.pool.QueryRow(ctx, query, name).Scan(&p.ID, &p.Name, &p.Description, &p.Resource, &p.Action, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrPermissionNotFound
	}
	return &p, err
}

func (r *permissionRepo) List(ctx context.Context) ([]domain.Permission, error) {
	query := `SELECT id, name, description, resource, action, created_at FROM permissions ORDER BY resource, action`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []domain.Permission
	for rows.Next() {
		var p domain.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Resource, &p.Action, &p.CreatedAt); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, nil
}

func (r *permissionRepo) ListByResource(ctx context.Context, resource string) ([]domain.Permission, error) {
	query := `SELECT id, name, description, resource, action, created_at FROM permissions WHERE resource = $1 ORDER BY action`
	rows, err := r.pool.Query(ctx, query, resource)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []domain.Permission
	for rows.Next() {
		var p domain.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Resource, &p.Action, &p.CreatedAt); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, nil
}

func (r *permissionRepo) Update(ctx context.Context, perm *domain.Permission) error {
	query := `UPDATE permissions SET name = $2, description = $3, resource = $4, action = $5 WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, perm.ID, perm.Name, perm.Description, perm.Resource, perm.Action)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrPermissionNotFound
	}
	return nil
}

func (r *permissionRepo) Delete(ctx context.Context, id domain.PermissionID) error {
	query := `DELETE FROM permissions WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrPermissionNotFound
	}
	return nil
}

var _ repository.PermissionRepository = (*permissionRepo)(nil)