package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cvsz/cfc-bot-manager/backend/internal/domain"
	"github.com/cvsz/cfc-bot-manager/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type organizationRepo struct {
	pool *pgxpool.Pool
}

func NewOrganizationRepository(pool *pgxpool.Pool) *organizationRepo {
	return &organizationRepo{pool: pool}
}

func (r *organizationRepo) Create(ctx context.Context, org *domain.Organization) error {
	query := `
		INSERT INTO organizations (id, name, slug, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, query,
		org.ID, org.Name, org.Slug, org.Description, org.CreatedAt, org.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create organization: %w", err)
	}
	return nil
}

func (r *organizationRepo) GetByID(ctx context.Context, id domain.OrganizationID) (*domain.Organization, error) {
	query := `
		SELECT id, name, slug, description, created_at, updated_at
		FROM organizations WHERE id = $1
	`
	var org domain.Organization
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&org.ID, &org.Name, &org.Slug, &org.Description, &org.CreatedAt, &org.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrOrganizationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get organization: %w", err)
	}
	return &org, nil
}

func (r *organizationRepo) GetBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	query := `
		SELECT id, name, slug, description, created_at, updated_at
		FROM organizations WHERE slug = $1
	`
	var org domain.Organization
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&org.ID, &org.Name, &org.Slug, &org.Description, &org.CreatedAt, &org.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrOrganizationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get organization by slug: %w", err)
	}
	return &org, nil
}

func (r *organizationRepo) List(ctx context.Context, pagination repository.Pagination, sort repository.Sort, filter repository.Filter) (repository.ListResult[domain.Organization], error) {
	// Build WHERE clause from filter
	whereClause, args := buildFilterClause(filter, 1)
	argIndex := len(args) + 1

	// Validate sort field
	sortField := "created_at"
	if sort.Field != "" {
		allowedSortFields := map[string]bool{"name": true, "slug": true, "created_at": true, "updated_at": true}
		if allowedSortFields[sort.Field] {
			sortField = sort.Field
		}
	}
	sortOrder := "DESC"
	if sort.Order == repository.SortAsc {
		sortOrder = "ASC"
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM organizations" + whereClause
	var totalCount int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return repository.ListResult[domain.Organization]{}, fmt.Errorf("count organizations: %w", err)
	}

	// Build data query
	dataQuery := fmt.Sprintf(`
		SELECT id, name, slug, description, created_at, updated_at
		FROM organizations
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortField, sortOrder, argIndex, argIndex+1)

	args = append(args, pagination.Limit, pagination.Offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return repository.ListResult[domain.Organization]{}, fmt.Errorf("list organizations: %w", err)
	}
	defer rows.Close()

	var orgs []domain.Organization
	for rows.Next() {
		var org domain.Organization
		if err := rows.Scan(&org.ID, &org.Name, &org.Slug, &org.Description, &org.CreatedAt, &org.UpdatedAt); err != nil {
			return repository.ListResult[domain.Organization]{}, fmt.Errorf("scan organization: %w", err)
		}
		orgs = append(orgs, org)
	}

	return repository.ListResult[domain.Organization]{
		Items:      orgs,
		TotalCount: totalCount,
		Pagination: pagination,
	}, nil
}

func (r *organizationRepo) Update(ctx context.Context, org *domain.Organization) error {
	query := `
		UPDATE organizations
		SET name = $2, slug = $3, description = $4, updated_at = $5
		WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query, org.ID, org.Name, org.Slug, org.Description, org.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update organization: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrOrganizationNotFound
	}
	return nil
}

func (r *organizationRepo) Delete(ctx context.Context, id domain.OrganizationID) error {
	query := `DELETE FROM organizations WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete organization: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrOrganizationNotFound
	}
	return nil
}

func (r *organizationRepo) ExistsBySlug(ctx context.Context, slug string, excludeID *domain.OrganizationID) (bool, error) {
	query := `SELECT 1 FROM organizations WHERE slug = $1`
	args := []any{slug}
	if excludeID != nil {
		query += ` AND id != $2`
		args = append(args, *excludeID)
	}
	var exists int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check slug exists: %w", err)
	}
	return true, nil
}

// Helper to build WHERE clause from filter
func buildFilterClause(filter repository.Filter, startIndex int) (string, []any) {
	if len(filter) == 0 {
		return "", nil
	}

	var conditions []string
	var args []any
	argIndex := startIndex

	for key, value := range filter {
		conditions = append(conditions, fmt.Sprintf("%s = $%d", key, argIndex))
		args = append(args, value)
		argIndex++
	}

	return " WHERE " + joinConditions(conditions), args
}

func joinConditions(conditions []string) string {
	result := ""
	for i, c := range conditions {
		if i > 0 {
			result += " AND "
		}
		result += c
	}
	return result
}

// Ensure pgtype.UUID can be scanned from UUID
func scanUUID(rows interface{ Scan(...any) error }, dest *domain.OrganizationID) error {
	var uuidVal pgtype.UUID
	if err := rows.Scan(&uuidVal); err != nil {
		return err
	}
	if uuidVal.Valid {
		*dest = domain.OrganizationID(uuidVal.Bytes)
	}
	return nil
}

// Need to fix the scanUUID function - it's not used correctly. Let me use a simpler approach.
// Actually, the domain IDs are just uuid.UUID aliases, so we can scan directly.

var _ repository.OrganizationRepository = (*organizationRepo)(nil)