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

type botConnectionRepo struct {
	pool *pgxpool.Pool
}

func NewBotConnectionRepository(pool *pgxpool.Pool) *botConnectionRepo {
	return &botConnectionRepo{pool: pool}
}

func (r *botConnectionRepo) Create(ctx context.Context, conn *domain.BotConnection) error {
	query := `
		INSERT INTO bot_connections (id, bot_id, room_id, event_type, message, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, query,
		conn.ID, conn.BotID, conn.RoomID, conn.EventType,
		conn.Message, conn.Metadata, conn.CreatedAt,
	)
	return err
}

func (r *botConnectionRepo) ListByBot(ctx context.Context, botID domain.BotID, pagination repository.Pagination) (repository.ListResult[domain.BotConnection], error) {
	countQuery := `SELECT COUNT(*) FROM bot_connections WHERE bot_id = $1`
	var totalCount int64
	err := r.pool.QueryRow(ctx, countQuery, botID).Scan(&totalCount)
	if err != nil {
		return repository.ListResult[domain.BotConnection]{}, err
	}

	query := `
		SELECT id, bot_id, room_id, event_type, message, metadata, created_at
		FROM bot_connections WHERE bot_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, botID, pagination.Limit, pagination.Offset)
	if err != nil {
		return repository.ListResult[domain.BotConnection]{}, err
	}
	defer rows.Close()

	var conns []domain.BotConnection
	for rows.Next() {
		var conn domain.BotConnection
		var roomID sql.NullString
		if err := rows.Scan(
			&conn.ID, &conn.BotID, &roomID, &conn.EventType,
			&conn.Message, &conn.Metadata, &conn.CreatedAt,
		); err != nil {
			return repository.ListResult[domain.BotConnection]{}, err
		}
		if roomID.Valid {
			rid := domain.RoomID(uuid.MustParse(roomID.String))
			conn.RoomID = &rid
		}
		conns = append(conns, conn)
	}

	return repository.ListResult[domain.BotConnection]{
		Items:      conns,
		TotalCount: totalCount,
		Pagination: pagination,
	}, nil
}

func (r *botConnectionRepo) ListByRoom(ctx context.Context, roomID domain.RoomID, pagination repository.Pagination) (repository.ListResult[domain.BotConnection], error) {
	countQuery := `SELECT COUNT(*) FROM bot_connections WHERE room_id = $1`
	var totalCount int64
	err := r.pool.QueryRow(ctx, countQuery, roomID).Scan(&totalCount)
	if err != nil {
		return repository.ListResult[domain.BotConnection]{}, err
	}

	query := `
		SELECT id, bot_id, room_id, event_type, message, metadata, created_at
		FROM bot_connections WHERE room_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, roomID, pagination.Limit, pagination.Offset)
	if err != nil {
		return repository.ListResult[domain.BotConnection]{}, err
	}
	defer rows.Close()

	var conns []domain.BotConnection
	for rows.Next() {
		var conn domain.BotConnection
		var rID sql.NullString
		if err := rows.Scan(
			&conn.ID, &conn.BotID, &rID, &conn.EventType,
			&conn.Message, &conn.Metadata, &conn.CreatedAt,
		); err != nil {
			return repository.ListResult[domain.BotConnection]{}, err
		}
		if rID.Valid {
			conn.RoomID = &roomID
		}
		conns = append(conns, conn)
	}

	return repository.ListResult[domain.BotConnection]{
		Items:      conns,
		TotalCount: totalCount,
		Pagination: pagination,
	}, nil
}

func (r *botConnectionRepo) DeleteOlderThan(ctx context.Context, olderThan time.Time) (int64, error) {
	query := `DELETE FROM bot_connections WHERE created_at < $1`
	result, err := r.pool.Exec(ctx, query, olderThan)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

var _ repository.BotConnectionRepository = (*botConnectionRepo)(nil)