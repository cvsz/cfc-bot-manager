package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/cvsz/cfc-bot-manager/backend/internal/domain"
	"github.com/cvsz/cfc-bot-manager/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgtype"
)

type botRepo struct {
	pool *pgxpool.Pool
}

func NewBotRepository(pool *pgxpool.Pool) *botRepo {
	return &botRepo{pool: pool}
}

func (r *botRepo) Create(ctx context.Context, bot *domain.Bot) error {
	query := `
		INSERT INTO bots (id, organization_id, name, username, password_hash, status, room_id, last_connected_at, last_error, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.pool.Exec(ctx, query,
		bot.ID, bot.OrganizationID, bot.Name, bot.Username, bot.PasswordHash,
		bot.Status, bot.RoomID, bot.LastConnectedAt, bot.LastError,
		bot.Config, bot.CreatedAt, bot.UpdatedAt,
	)
	return err
}

func (r *botRepo) GetByID(ctx context.Context, id domain.BotID) (*domain.Bot, error) {
	query := `
		SELECT id, organization_id, name, username, password_hash, status, room_id, last_connected_at, last_error, config, created_at, updated_at
		FROM bots WHERE id = $1
	`
	var bot domain.Bot
	var roomID pgtype.UUID
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&bot.ID, &bot.OrganizationID, &bot.Name, &bot.Username, &bot.PasswordHash,
		&bot.Status, &roomID, &bot.LastConnectedAt, &bot.LastError,
		&bot.Config, &bot.CreatedAt, &bot.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrBotNotFound
	}
	if roomID.Valid {
		rid := domain.RoomID(roomID.Bytes)
		bot.RoomID = &rid
	}
	return &bot, err
}

func (r *botRepo) GetByUsername(ctx context.Context, orgID domain.OrganizationID, username string) (*domain.Bot, error) {
	query := `
		SELECT id, organization_id, name, username, password_hash, status, room_id, last_connected_at, last_error, config, created_at, updated_at
		FROM bots WHERE organization_id = $1 AND username = $2
	`
	var bot domain.Bot
	var roomID pgtype.UUID
	err := r.pool.QueryRow(ctx, query, orgID, username).Scan(
		&bot.ID, &bot.OrganizationID, &bot.Name, &bot.Username, &bot.PasswordHash,
		&bot.Status, &roomID, &bot.LastConnectedAt, &bot.LastError,
		&bot.Config, &bot.CreatedAt, &bot.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrBotNotFound
	}
	if roomID.Valid {
		rid := domain.RoomID(roomID.Bytes)
		bot.RoomID = &rid
	}
	return &bot, err
}

func (r *botRepo) ListByOrg(ctx context.Context, orgID domain.OrganizationID, pagination repository.Pagination, sort repository.Sort, filter repository.Filter) (repository.ListResult[domain.Bot], error) {
	whereClause, args := buildFilterClauseWithOrg(filter, orgID, 1)
	argIndex := len(args) + 1

	sortField := "created_at"
	if sort.Field != "" {
		allowedSortFields := map[string]bool{"name": true, "username": true, "status": true, "created_at": true, "updated_at": true}
		if allowedSortFields[sort.Field] {
			sortField = sort.Field
		}
	}
	sortOrder := "DESC"
	if sort.Order == repository.SortAsc {
		sortOrder = "ASC"
	}

	countQuery := "SELECT COUNT(*) FROM bots" + whereClause
	var totalCount int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return repository.ListResult[domain.Bot]{}, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, organization_id, name, username, password_hash, status, room_id, last_connected_at, last_error, config, created_at, updated_at
		FROM bots
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortField, sortOrder, argIndex, argIndex+1)

	args = append(args, pagination.Limit, pagination.Offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return repository.ListResult[domain.Bot]{}, err
	}
	defer rows.Close()

	var bots []domain.Bot
	for rows.Next() {
		var bot domain.Bot
		var roomID pgtype.UUID
		if err := rows.Scan(
			&bot.ID, &bot.OrganizationID, &bot.Name, &bot.Username, &bot.PasswordHash,
			&bot.Status, &roomID, &bot.LastConnectedAt, &bot.LastError,
			&bot.Config, &bot.CreatedAt, &bot.UpdatedAt,
		); err != nil {
			return repository.ListResult[domain.Bot]{}, err
		}
		if roomID.Valid {
			rid := domain.RoomID(roomID.Bytes)
			bot.RoomID = &rid
		}
		bots = append(bots, bot)
	}

	return repository.ListResult[domain.Bot]{
		Items:      bots,
		TotalCount: totalCount,
		Pagination: pagination,
	}, nil
}

func (r *botRepo) ListByRoom(ctx context.Context, roomID domain.RoomID) ([]domain.Bot, error) {
	query := `
		SELECT id, organization_id, name, username, password_hash, status, room_id, last_connected_at, last_error, config, created_at, updated_at
		FROM bots WHERE room_id = $1
	`
	rows, err := r.pool.Query(ctx, query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bots []domain.Bot
	for rows.Next() {
		var bot domain.Bot
		var rid pgtype.UUID
		if err := rows.Scan(
			&bot.ID, &bot.OrganizationID, &bot.Name, &bot.Username, &bot.PasswordHash,
			&bot.Status, &rid, &bot.LastConnectedAt, &bot.LastError,
			&bot.Config, &bot.CreatedAt, &bot.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if rid.Valid {
			bot.RoomID = &roomID
		}
		bots = append(bots, bot)
	}
	return bots, nil
}

func (r *botRepo) Update(ctx context.Context, bot *domain.Bot) error {
	query := `
		UPDATE bots
		SET name = $2, username = $3, password_hash = $4, status = $5, room_id = $6, last_connected_at = $7, last_error = $8, config = $9, updated_at = $10
		WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query,
		bot.ID, bot.Name, bot.Username, bot.PasswordHash, bot.Status,
		bot.RoomID, bot.LastConnectedAt, bot.LastError, bot.Config, bot.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrBotNotFound
	}
	return nil
}

func (r *botRepo) UpdateStatus(ctx context.Context, id domain.BotID, status string, lastError string, lastConnected *time.Time) error {
	query := `UPDATE bots SET status = $1, last_error = $2, last_connected_at = $3, updated_at = $4 WHERE id = $5`
	_, err := r.pool.Exec(ctx, query, status, lastError, lastConnected, time.Now(), id)
	return err
}

func (r *botRepo) UpdateConfig(ctx context.Context, id domain.BotID, config domain.BotConfig) error {
	query := `UPDATE bots SET config = $1, updated_at = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, config, time.Now(), id)
	return err
}

func (r *botRepo) AssignRoom(ctx context.Context, botID domain.BotID, roomID *domain.RoomID) error {
	query := `UPDATE bots SET room_id = $1, updated_at = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, roomID, time.Now(), botID)
	return err
}

func (r *botRepo) Delete(ctx context.Context, id domain.BotID) error {
	query := `DELETE FROM bots WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrBotNotFound
	}
	return nil
}

func (r *botRepo) ExistsByUsername(ctx context.Context, orgID domain.OrganizationID, username string, excludeID *domain.BotID) (bool, error) {
	query := `SELECT 1 FROM bots WHERE organization_id = $1 AND username = $2`
	args := []any{orgID, username}
	if excludeID != nil {
		query += ` AND id != $3`
		args = append(args, *excludeID)
	}
	var exists int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return true, err
}

var _ repository.BotRepository = (*botRepo)(nil)