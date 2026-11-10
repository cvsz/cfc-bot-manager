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

type roomRepo struct {
	pool *pgxpool.Pool
}

func NewRoomRepository(pool *pgxpool.Pool) *roomRepo {
	return &roomRepo{pool: pool}
}

func (r *roomRepo) Create(ctx context.Context, room *domain.Room) error {
	query := `
		INSERT INTO rooms (id, organization_id, name, access_key_hash, bot_id, language, is_paused, status, last_seen_at, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.pool.Exec(ctx, query,
		room.ID, room.OrganizationID, room.Name, room.AccessKeyHash,
		room.BotID, room.Language, room.IsPaused, room.Status,
		room.LastSeenAt, room.Config, room.CreatedAt, room.UpdatedAt,
	)
	return err
}

func (r *roomRepo) GetByID(ctx context.Context, id domain.RoomID) (*domain.Room, error) {
	query := `
		SELECT id, organization_id, name, access_key_hash, bot_id, language, is_paused, status, last_seen_at, config, created_at, updated_at
		FROM rooms WHERE id = $1
	`
	var room domain.Room
	var botID pgtype.UUID
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&room.ID, &room.OrganizationID, &room.Name, &room.AccessKeyHash,
		&botID, &room.Language, &room.IsPaused, &room.Status,
		&room.LastSeenAt, &room.Config, &room.CreatedAt, &room.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrRoomNotFound
	}
	if botID.Valid {
		bid := domain.BotID(botID.Bytes)
		room.BotID = &bid
	}
	return &room, err
}

func (r *roomRepo) GetByOrgAndName(ctx context.Context, orgID domain.OrganizationID, name string) (*domain.Room, error) {
	query := `
		SELECT id, organization_id, name, access_key_hash, bot_id, language, is_paused, status, last_seen_at, config, created_at, updated_at
		FROM rooms WHERE organization_id = $1 AND name = $2
	`
	var room domain.Room
	var botID pgtype.UUID
	err := r.pool.QueryRow(ctx, query, orgID, name).Scan(
		&room.ID, &room.OrganizationID, &room.Name, &room.AccessKeyHash,
		&botID, &room.Language, &room.IsPaused, &room.Status,
		&room.LastSeenAt, &room.Config, &room.CreatedAt, &room.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrRoomNotFound
	}
	if botID.Valid {
		bid := domain.BotID(botID.Bytes)
		room.BotID = &bid
	}
	return &room, err
}

func (r *roomRepo) ListByOrg(ctx context.Context, orgID domain.OrganizationID, pagination repository.Pagination, sort repository.Sort, filter repository.Filter) (repository.ListResult[domain.Room], error) {
	whereClause, args := buildFilterClauseWithOrg(filter, orgID, 1)
	argIndex := len(args) + 1

	sortField := "created_at"
	if sort.Field != "" {
		allowedSortFields := map[string]bool{"name": true, "created_at": true, "updated_at": true, "status": true, "is_paused": true}
		if allowedSortFields[sort.Field] {
			sortField = sort.Field
		}
	}
	sortOrder := "DESC"
	if sort.Order == repository.SortAsc {
		sortOrder = "ASC"
	}

	countQuery := "SELECT COUNT(*) FROM rooms" + whereClause
	var totalCount int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return repository.ListResult[domain.Room]{}, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, organization_id, name, access_key_hash, bot_id, language, is_paused, status, last_seen_at, config, created_at, updated_at
		FROM rooms
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortField, sortOrder, argIndex, argIndex+1)

	args = append(args, pagination.Limit, pagination.Offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return repository.ListResult[domain.Room]{}, err
	}
	defer rows.Close()

	var rooms []domain.Room
	for rows.Next() {
		var room domain.Room
		var botID pgtype.UUID
		if err := rows.Scan(
			&room.ID, &room.OrganizationID, &room.Name, &room.AccessKeyHash,
			&botID, &room.Language, &room.IsPaused, &room.Status,
			&room.LastSeenAt, &room.Config, &room.CreatedAt, &room.UpdatedAt,
		); err != nil {
			return repository.ListResult[domain.Room]{}, err
		}
		if botID.Valid {
			bid := domain.BotID(botID.Bytes)
			room.BotID = &bid
		}
		rooms = append(rooms, room)
	}

	return repository.ListResult[domain.Room]{
		Items:      rooms,
		TotalCount: totalCount,
		Pagination: pagination,
	}, nil
}

func (r *roomRepo) ListByBot(ctx context.Context, botID domain.BotID) ([]domain.Room, error) {
	query := `
		SELECT id, organization_id, name, access_key_hash, bot_id, language, is_paused, status, last_seen_at, config, created_at, updated_at
		FROM rooms WHERE bot_id = $1
	`
	rows, err := r.pool.Query(ctx, query, botID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []domain.Room
	for rows.Next() {
		var room domain.Room
		var bid pgtype.UUID
		if err := rows.Scan(
			&room.ID, &room.OrganizationID, &room.Name, &room.AccessKeyHash,
			&bid, &room.Language, &room.IsPaused, &room.Status,
			&room.LastSeenAt, &room.Config, &room.CreatedAt, &room.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if bid.Valid {
			room.BotID = &botID
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func (r *roomRepo) Update(ctx context.Context, room *domain.Room) error {
	query := `
		UPDATE rooms
		SET name = $2, access_key_hash = $3, bot_id = $4, language = $5, is_paused = $6, status = $7, last_seen_at = $8, config = $9, updated_at = $10
		WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query,
		room.ID, room.Name, room.AccessKeyHash, room.BotID,
		room.Language, room.IsPaused, room.Status, room.LastSeenAt,
		room.Config, room.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrRoomNotFound
	}
	return nil
}

func (r *roomRepo) UpdateStatus(ctx context.Context, id domain.RoomID, status string, lastSeen *time.Time) error {
	query := `UPDATE rooms SET status = $1, last_seen_at = $2, updated_at = $3 WHERE id = $4`
	_, err := r.pool.Exec(ctx, query, status, lastSeen, time.Now(), id)
	return err
}

func (r *roomRepo) UpdateConfig(ctx context.Context, id domain.RoomID, config domain.RoomConfig) error {
	query := `UPDATE rooms SET config = $1, updated_at = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, config, time.Now(), id)
	return err
}

func (r *roomRepo) AssignBot(ctx context.Context, roomID domain.RoomID, botID *domain.BotID) error {
	query := `UPDATE rooms SET bot_id = $1, updated_at = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, botID, time.Now(), roomID)
	return err
}

func (r *roomRepo) Delete(ctx context.Context, id domain.RoomID) error {
	query := `DELETE FROM rooms WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrRoomNotFound
	}
	return nil
}

func (r *roomRepo) ExistsByName(ctx context.Context, orgID domain.OrganizationID, name string, excludeID *domain.RoomID) (bool, error) {
	query := `SELECT 1 FROM rooms WHERE organization_id = $1 AND name = $2`
	args := []any{orgID, name}
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

var _ repository.RoomRepository = (*roomRepo)(nil)