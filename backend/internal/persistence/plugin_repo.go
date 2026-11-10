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

type pluginRepo struct {
	pool *pgxpool.Pool
}

func NewPluginRepository(pool *pgxpool.Pool) *pluginRepo {
	return &pluginRepo{pool: pool}
}

func (r *pluginRepo) Create(ctx context.Context, plugin *domain.Plugin) error {
	query := `
		INSERT INTO plugins (id, plugin_number, name, description, version, compatibility, capability_flags, config_schema, is_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.pool.Exec(ctx, query,
		plugin.ID, plugin.PluginNumber, plugin.Name, plugin.Description,
		plugin.Version, plugin.Compatibility, plugin.CapabilityFlags,
		plugin.ConfigSchema, plugin.IsEnabled, plugin.CreatedAt, plugin.UpdatedAt,
	)
	return err
}

func (r *pluginRepo) GetByID(ctx context.Context, id domain.PluginID) (*domain.Plugin, error) {
	query := `
		SELECT id, plugin_number, name, description, version, compatibility, capability_flags, config_schema, is_enabled, created_at, updated_at
		FROM plugins WHERE id = $1
	`
	var plugin domain.Plugin
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&plugin.ID, &plugin.PluginNumber, &plugin.Name, &plugin.Description,
		&plugin.Version, &plugin.Compatibility, &plugin.CapabilityFlags,
		&plugin.ConfigSchema, &plugin.IsEnabled, &plugin.CreatedAt, &plugin.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrPluginNotFound
	}
	return &plugin, err
}

func (r *pluginRepo) GetByNumber(ctx context.Context, number int) (*domain.Plugin, error) {
	query := `
		SELECT id, plugin_number, name, description, version, compatibility, capability_flags, config_schema, is_enabled, created_at, updated_at
		FROM plugins WHERE plugin_number = $1
	`
	var plugin domain.Plugin
	err := r.pool.QueryRow(ctx, query, number).Scan(
		&plugin.ID, &plugin.PluginNumber, &plugin.Name, &plugin.Description,
		&plugin.Version, &plugin.Compatibility, &plugin.CapabilityFlags,
		&plugin.ConfigSchema, &plugin.IsEnabled, &plugin.CreatedAt, &plugin.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrPluginNotFound
	}
	return &plugin, err
}

func (r *pluginRepo) List(ctx context.Context, pagination repository.Pagination, sort repository.Sort, filter repository.Filter) (repository.ListResult[domain.Plugin], error) {
	whereClause, args := buildFilterClause(filter, 1)
	argIndex := len(args) + 1

	sortField := "plugin_number"
	if sort.Field != "" {
		allowedSortFields := map[string]bool{"plugin_number": true, "name": true, "compatibility": true, "is_enabled": true}
		if allowedSortFields[sort.Field] {
			sortField = sort.Field
		}
	}
	sortOrder := "ASC"
	if sort.Order == repository.SortAsc {
		sortOrder = "ASC"
	} else {
		sortOrder = "DESC"
	}

	countQuery := "SELECT COUNT(*) FROM plugins" + whereClause
	var totalCount int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return repository.ListResult[domain.Plugin]{}, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, plugin_number, name, description, version, compatibility, capability_flags, config_schema, is_enabled, created_at, updated_at
		FROM plugins
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortField, sortOrder, argIndex, argIndex+1)

	args = append(args, pagination.Limit, pagination.Offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return repository.ListResult[domain.Plugin]{}, err
	}
	defer rows.Close()

	var plugins []domain.Plugin
	for rows.Next() {
		var plugin domain.Plugin
		if err := rows.Scan(
			&plugin.ID, &plugin.PluginNumber, &plugin.Name, &plugin.Description,
			&plugin.Version, &plugin.Compatibility, &plugin.CapabilityFlags,
			&plugin.ConfigSchema, &plugin.IsEnabled, &plugin.CreatedAt, &plugin.UpdatedAt,
		); err != nil {
			return repository.ListResult[domain.Plugin]{}, err
		}
		plugins = append(plugins, plugin)
	}

	return repository.ListResult[domain.Plugin]{
		Items:      plugins,
		TotalCount: totalCount,
		Pagination: pagination,
	}, nil
}

func (r *pluginRepo) ListEnabled(ctx context.Context) ([]domain.Plugin, error) {
	query := `
		SELECT id, plugin_number, name, description, version, compatibility, capability_flags, config_schema, is_enabled, created_at, updated_at
		FROM plugins WHERE is_enabled = TRUE ORDER BY plugin_number
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plugins []domain.Plugin
	for rows.Next() {
		var plugin domain.Plugin
		if err := rows.Scan(
			&plugin.ID, &plugin.PluginNumber, &plugin.Name, &plugin.Description,
			&plugin.Version, &plugin.Compatibility, &plugin.CapabilityFlags,
			&plugin.ConfigSchema, &plugin.IsEnabled, &plugin.CreatedAt, &plugin.UpdatedAt,
		); err != nil {
			return nil, err
		}
		plugins = append(plugins, plugin)
	}
	return plugins, nil
}

func (r *pluginRepo) Update(ctx context.Context, plugin *domain.Plugin) error {
	query := `
		UPDATE plugins
		SET name = $2, description = $3, version = $4, compatibility = $5, capability_flags = $6, config_schema = $7, is_enabled = $8, updated_at = $9
		WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query,
		plugin.ID, plugin.Name, plugin.Description, plugin.Version,
		plugin.Compatibility, plugin.CapabilityFlags, plugin.ConfigSchema,
		plugin.IsEnabled, plugin.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrPluginNotFound
	}
	return nil
}

func (r *pluginRepo) UpdateCompatibility(ctx context.Context, id domain.PluginID, compat string) error {
	query := `UPDATE plugins SET compatibility = $1, updated_at = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, compat, time.Now(), id)
	return err
}

func (r *pluginRepo) Delete(ctx context.Context, id domain.PluginID) error {
	query := `DELETE FROM plugins WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrPluginNotFound
	}
	return nil
}

func (r *pluginRepo) ExistsByNumber(ctx context.Context, number int, excludeID *domain.PluginID) (bool, error) {
	query := `SELECT 1 FROM plugins WHERE plugin_number = $1`
	args := []any{number}
	if excludeID != nil {
		query += ` AND id != $2`
		args = append(args, *excludeID)
	}
	var exists int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return true, err
}

var _ repository.PluginRepository = (*pluginRepo)(nil)