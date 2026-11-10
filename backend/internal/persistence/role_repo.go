package persistence

import (
	"context"
	"database/sql"

	"github.com/cvsz/cfc-bot-manager/backend/internal/domain"
	"github.com/cvsz/cfc-bot-manager/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type roleRepo struct {
	pool *pgxpool.Pool
}

func NewRoleRepository(pool *pgxpool.Pool) *roleRepo {
	return &roleRepo{pool: pool}
}

func (r *roleRepo) Create(ctx context.Context, role *domain.Role) error {
	query := `
		INSERT INTO roles (id, organization_id, name, description, is_system, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, query, role.ID, role.OrganizationID, role.Name, role.Description, role.IsSystem, role.CreatedAt, role.UpdatedAt)
	return err
}

func (r *roleRepo) GetByID(ctx context.Context, id domain.RoleID) (*domain.Role, error) {
	query := `SELECT id, organization_id, name, description, is_system, created_at, updated_at FROM roles WHERE id = $1`
	var role domain.Role
	err := r.pool.QueryRow(ctx, query, id).Scan(&role.ID, &role.OrganizationID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrRoleNotFound
	}
	return &role, err
}

func (r *roleRepo) GetByName(ctx context.Context, orgID domain.OrganizationID, name string) (*domain.Role, error) {
	query := `SELECT id, organization_id, name, description, is_system, created_at, updated_at FROM roles WHERE organization_id = $1 AND name = $2`
	var role domain.Role
	err := r.pool.QueryRow(ctx, query, orgID, name).Scan(&role.ID, &role.OrganizationID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrRoleNotFound
	}
	return &role, err
}

func (r *roleRepo) ListByOrg(ctx context.Context, orgID domain.OrganizationID, includeSystem bool) ([]domain.Role, error) {
	query := `SELECT id, organization_id, name, description, is_system, created_at, updated_at FROM roles WHERE organization_id = $1`
	if !includeSystem {
		query += ` AND is_system = FALSE`
	}
	query += ` ORDER BY name`
	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []domain.Role
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.ID, &role.OrganizationID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *roleRepo) Update(ctx context.Context, role *domain.Role) error {
	query := `UPDATE roles SET name = $2, description = $3, updated_at = $4 WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, role.ID, role.Name, role.Description, role.UpdatedAt)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrRoleNotFound
	}
	return nil
}

func (r *roleRepo) Delete(ctx context.Context, id domain.RoleID) error {
	// Check if system role
	var isSystem bool
	err := r.pool.QueryRow(ctx, `SELECT is_system FROM roles WHERE id = $1`, id).Scan(&isSystem)
	if err == sql.ErrNoRows {
		return domain.ErrRoleNotFound
	}
	if err != nil {
		return err
	}
	if isSystem {
		return domain.ErrSystemRoleImmutable
	}

	query := `DELETE FROM roles WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrRoleNotFound
	}
	return nil
}

func (r *roleRepo) AddPermission(ctx context.Context, roleID domain.RoleID, permID domain.PermissionID) error {
	query := `INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.pool.Exec(ctx, query, roleID, permID)
	return err
}

func (r *roleRepo) RemovePermission(ctx context.Context, roleID domain.RoleID, permID domain.PermissionID) error {
	query := `DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2`
	_, err := r.pool.Exec(ctx, query, roleID, permID)
	return err
}

func (r *roleRepo) ListPermissions(ctx context.Context, roleID domain.RoleID) ([]domain.Permission, error) {
	query := `
		SELECT p.id, p.name, p.description, p.resource, p.action, p.created_at
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.resource, p.action
	`
	rows, err := r.pool.Query(ctx, query, roleID)
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

func (r *roleRepo) IsSystemRole(ctx context.Context, id domain.RoleID) (bool, error) {
	var isSystem bool
	err := r.pool.QueryRow(ctx, `SELECT is_system FROM roles WHERE id = $1`, id).Scan(&isSystem)
	if err == sql.ErrNoRows {
		return false, domain.ErrRoleNotFound
	}
	return isSystem, err
}

var _ repository.RoleRepository = (*roleRepo)(nil)