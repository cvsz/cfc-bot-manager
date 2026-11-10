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

type membershipRepo struct {
	pool *pgxpool.Pool
}

func NewMembershipRepository(pool *pgxpool.Pool) *membershipRepo {
	return &membershipRepo{pool: pool}
}

func (r *membershipRepo) Create(ctx context.Context, m *domain.Membership) error {
	query := `INSERT INTO memberships (id, user_id, organization_id, role_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.pool.Exec(ctx, query, m.ID, m.UserID, m.OrganizationID, m.RoleID, m.CreatedAt, m.UpdatedAt)
	return err
}

func (r *membershipRepo) GetByUserAndOrg(ctx context.Context, userID domain.UserID, orgID domain.OrganizationID) (*domain.Membership, error) {
	query := `SELECT id, user_id, organization_id, role_id, created_at, updated_at FROM memberships WHERE user_id = $1 AND organization_id = $2`
	var m domain.Membership
	err := r.pool.QueryRow(ctx, query, userID, orgID).Scan(&m.ID, &m.UserID, &m.OrganizationID, &m.RoleID, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrMembershipNotFound
	}
	return &m, err
}

func (r *membershipRepo) ListByUser(ctx context.Context, userID domain.UserID) ([]domain.Membership, error) {
	query := `SELECT id, user_id, organization_id, role_id, created_at, updated_at FROM memberships WHERE user_id = $1`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memberships []domain.Membership
	for rows.Next() {
		var m domain.Membership
		if err := rows.Scan(&m.ID, &m.UserID, &m.OrganizationID, &m.RoleID, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		memberships = append(memberships, m)
	}
	return memberships, nil
}

func (r *membershipRepo) ListByOrg(ctx context.Context, orgID domain.OrganizationID, pagination repository.Pagination, sort repository.Sort) (repository.ListResult[domain.Membership], error) {
	sortField := "created_at"
	if sort.Field != "" {
		allowedSortFields := map[string]bool{"created_at": true, "updated_at": true}
		if allowedSortFields[sort.Field] {
			sortField = sort.Field
		}
	}
	sortOrder := "DESC"
	if sort.Order == repository.SortAsc {
		sortOrder = "ASC"
	}

	countQuery := `SELECT COUNT(*) FROM memberships WHERE organization_id = $1`
	var totalCount int64
	err := r.pool.QueryRow(ctx, countQuery, orgID).Scan(&totalCount)
	if err != nil {
		return repository.ListResult[domain.Membership]{}, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, user_id, organization_id, role_id, created_at, updated_at
		FROM memberships WHERE organization_id = $1
		ORDER BY %s %s
		LIMIT $2 OFFSET $3
	`, sortField, sortOrder)

	rows, err := r.pool.Query(ctx, dataQuery, orgID, pagination.Limit, pagination.Offset)
	if err != nil {
		return repository.ListResult[domain.Membership]{}, err
	}
	defer rows.Close()

	var memberships []domain.Membership
	for rows.Next() {
		var m domain.Membership
		if err := rows.Scan(&m.ID, &m.UserID, &m.OrganizationID, &m.RoleID, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return repository.ListResult[domain.Membership]{}, err
		}
		memberships = append(memberships, m)
	}

	return repository.ListResult[domain.Membership]{
		Items:      memberships,
		TotalCount: totalCount,
		Pagination: pagination,
	}, nil
}

func (r *membershipRepo) UpdateRole(ctx context.Context, userID domain.UserID, orgID domain.OrganizationID, roleID domain.RoleID) error {
	query := `UPDATE memberships SET role_id = $1, updated_at = $2 WHERE user_id = $3 AND organization_id = $4`
	result, err := r.pool.Exec(ctx, query, roleID, time.Now(), userID, orgID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrMembershipNotFound
	}
	return nil
}

func (r *membershipRepo) Delete(ctx context.Context, userID domain.UserID, orgID domain.OrganizationID) error {
	query := `DELETE FROM memberships WHERE user_id = $1 AND organization_id = $2`
	result, err := r.pool.Exec(ctx, query, userID, orgID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrMembershipNotFound
	}
	return nil
}

func (r *membershipRepo) CountByOrgAndRole(ctx context.Context, orgID domain.OrganizationID, roleID domain.RoleID) (int64, error) {
	query := `SELECT COUNT(*) FROM memberships WHERE organization_id = $1 AND role_id = $2`
	var count int64
	err := r.pool.QueryRow(ctx, query, orgID, roleID).Scan(&count)
	return count, err
}

var _ repository.MembershipRepository = (*membershipRepo)(nil)