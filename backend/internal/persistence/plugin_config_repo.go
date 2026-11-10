package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/cvsz/cfc-bot-manager/backend/internal/domain"
	"github.com/cvsz/cfc-bot-manager/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pluginConfigRepo struct {
	pool *pgxpool.Pool
}

func NewPluginConfigRepository(pool *pgxpool.Pool) *pluginConfigRepo {
	return &pluginConfigRepo{pool: pool}
}

func (r *pluginConfigRepo) Create(ctx context.Context, config *domain.PluginConfiguration) error {
	query := `
		INSERT INTO plugin_configurations (id, plugin_id, room_id, bot_id, settings, is_enabled, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.pool.Exec(ctx, query,
		config.ID, config.PluginID, config.RoomID, config.BotID,
		config.Settings, config.IsEnabled, config.Version,
		config.CreatedAt, config.UpdatedAt,
	)
	return err
}

func (r *pluginConfigRepo) GetByID(ctx context.Context, id domain.PluginConfigID) (*domain.PluginConfiguration, error) {
	query := `
		SELECT id, plugin_id, room_id, bot_id, settings, is_enabled, version, created_at, updated_at
		FROM plugin_configurations WHERE id = $1
	`
	var config domain.PluginConfiguration
	var roomID, botID sql.NullString // Using sql.NullString for nullable UUIDs
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&config.ID, &config.PluginID, &roomID, &botID,
		&config.Settings, &config.IsEnabled, &config.Version,
		&config.CreatedAt, &config.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrPluginConfigNotFound
	}
	if roomID.Valid {
		rid := domain.RoomID(uuid.MustParse(roomID.String))
		config.RoomID = &rid
	}
	if botID.Valid {
		bid := domain.BotID(uuid.MustParse(botID.String))
		config.BotID = &bid
	}
	return &config, err
}

func (r *pluginConfigRepo) GetByPluginAndRoom(ctx context.Context, pluginID domain.PluginID, roomID domain.RoomID) (*domain.PluginConfiguration, error) {
	query := `
		SELECT id, plugin_id, room_id, bot_id, settings, is_enabled, version, created_at, updated_at
		FROM plugin_configurations WHERE plugin_id = $1 AND room_id = $2
	`
	var config domain.PluginConfiguration
	var rID, bID sql.NullString
	err := r.pool.QueryRow(ctx, query, pluginID, roomID).Scan(
		&config.ID, &config.PluginID, &rID, &bID,
		&config.Settings, &config.IsEnabled, &config.Version,
		&config.CreatedAt, &config.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrPluginConfigNotFound
	}
	if rID.Valid {
		rid := domain.RoomID(uuid.MustParse(rID.String))
		config.RoomID = &rid
	}
	return &config, err
}

func (r *pluginConfigRepo) GetByPluginAndBot(ctx context.Context, pluginID domain.PluginID, botID domain.BotID) (*domain.PluginConfiguration, error) {
	query := `
		SELECT id, plugin_id, room_id, bot_id, settings, is_enabled, version, created_at, updated_at
		FROM plugin_configurations WHERE plugin_id = $1 AND bot_id = $2
	`
	var config domain.PluginConfiguration
	var rID, bID sql.NullString
	err := r.pool.QueryRow(ctx, query, pluginID, botID).Scan(
		&config.ID, &config.PluginID, &rID, &bID,
		&config.Settings, &config.IsEnabled, &config.Version,
		&config.CreatedAt, &config.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrPluginConfigNotFound
	}
	if bID.Valid {
		bid := domain.BotID(uuid.MustParse(bID.String))
		config.BotID = &bid
	}
	return &config, err
}

func (r *pluginConfigRepo) ListByRoom(ctx context.Context, roomID domain.RoomID) ([]domain.PluginConfiguration, error) {
	query := `
		SELECT id, plugin_id, room_id, bot_id, settings, is_enabled, version, created_at, updated_at
		FROM plugin_configurations WHERE room_id = $1
	`
	rows, err := r.pool.Query(ctx, query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []domain.PluginConfiguration
	for rows.Next() {
		var config domain.PluginConfiguration
		var rID, bID sql.NullString
		if err := rows.Scan(
			&config.ID, &config.PluginID, &rID, &bID,
			&config.Settings, &config.IsEnabled, &config.Version,
			&config.CreatedAt, &config.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if rID.Valid {
			rid := domain.RoomID(uuid.MustParse(rID.String))
			config.RoomID = &rid
		}
		configs = append(configs, config)
	}
	return configs, nil
}

func (r *pluginConfigRepo) ListByBot(ctx context.Context, botID domain.BotID) ([]domain.PluginConfiguration, error) {
	query := `
		SELECT id, plugin_id, room_id, bot_id, settings, is_enabled, version, created_at, updated_at
		FROM plugin_configurations WHERE bot_id = $1
	`
	rows, err := r.pool.Query(ctx, query, botID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []domain.PluginConfiguration
	for rows.Next() {
		var config domain.PluginConfiguration
		var rID, bID sql.NullString
		if err := rows.Scan(
			&config.ID, &config.PluginID, &rID, &bID,
			&config.Settings, &config.IsEnabled, &config.Version,
			&config.CreatedAt, &config.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if bID.Valid {
			bid := domain.BotID(uuid.MustParse(bID.String))
			config.BotID = &bid
		}
		configs = append(configs, config)
	}
	return configs, nil
}

func (r *pluginConfigRepo) Update(ctx context.Context, config *domain.PluginConfiguration) error {
	query := `
		UPDATE plugin_configurations
		SET plugin_id = $2, room_id = $3, bot_id = $4, settings = $5, is_enabled = $6, version = $7, updated_at = $8
		WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query,
		config.ID, config.PluginID, config.RoomID, config.BotID,
		config.Settings, config.IsEnabled, config.Version, config.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrPluginConfigNotFound
	}
	return nil
}

func (r *pluginConfigRepo) UpdateSettings(ctx context.Context, id domain.PluginConfigID, settings string, version int) error {
	query := `UPDATE plugin_configurations SET settings = $1, version = $2, updated_at = $3 WHERE id = $4 AND version = $5`
	result, err := r.pool.Exec(ctx, query, settings, version, time.Now(), id, version-1)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrOptimisticLock
	}
	return nil
}

func (r *pluginConfigRepo) ToggleEnabled(ctx context.Context, id domain.PluginConfigID, enabled bool) error {
	query := `UPDATE plugin_configurations SET is_enabled = $1, updated_at = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, enabled, time.Now(), id)
	return err
}

func (r *pluginConfigRepo) Delete(ctx context.Context, id domain.PluginConfigID) error {
	query := `DELETE FROM plugin_configurations WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrPluginConfigNotFound
	}
	return nil
}

func (r *pluginConfigRepo) DeleteByRoom(ctx context.Context, roomID domain.RoomID) error {
	query := `DELETE FROM plugin_configurations WHERE room_id = $1`
	_, err := r.pool.Exec(ctx, query, roomID)
	return err
}

func (r *pluginConfigRepo) DeleteByBot(ctx context.Context, botID domain.BotID) error {
	query := `DELETE FROM plugin_configurations WHERE bot_id = $1`
	_, err := r.pool.Exec(ctx, query, botID)
	return err
}

var _ repository.PluginConfigRepository = (*pluginConfigRepo)(nil)