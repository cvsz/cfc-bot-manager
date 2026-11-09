package messagestyle

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"strings"
)

//go:embed migrations/001_room_message_styles.sql
var initialMigration string

type Store interface {
	GetDefault(context.Context) (Style, error)
	SaveDefault(context.Context, Style) error
	GetRoom(context.Context, string) (Style, bool, error)
	SaveRoom(context.Context, string, Style) error
	DeleteRoom(context.Context, string) error
}

type PostgresStore struct{ db *sql.DB }

func NewPostgresStore(db *sql.DB) *PostgresStore { return &PostgresStore{db: db} }

func ApplyMigrations(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err = tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version integer PRIMARY KEY,
		applied_at timestamptz NOT NULL DEFAULT now()
	)`); err != nil {
		return fmt.Errorf("create migration table: %w", err)
	}
	if _, err = tx.ExecContext(ctx, "LOCK TABLE schema_migrations IN EXCLUSIVE MODE"); err != nil {
		return fmt.Errorf("lock migrations: %w", err)
	}
	var applied bool
	if err = tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = 1)").Scan(&applied); err != nil {
		return fmt.Errorf("read migration state: %w", err)
	}
	if !applied {
		for _, statement := range strings.Split(initialMigration, ";") {
			statement = strings.TrimSpace(statement)
			if statement == "" {
				continue
			}
			if _, err = tx.ExecContext(ctx, statement); err != nil {
				return fmt.Errorf("apply room style migration: %w", err)
			}
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES (1)"); err != nil {
			return fmt.Errorf("record migration: %w", err)
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit migration: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetDefault(ctx context.Context) (Style, error) {
	var v Style
	err := s.db.QueryRowContext(ctx, `SELECT font_family, font_size, color, bold, italic, underline
		FROM default_message_style WHERE singleton = TRUE`).Scan(
		&v.FontFamily, &v.FontSize, &v.Color, &v.Bold, &v.Italic, &v.Underline,
	)
	if err != nil {
		return Style{}, fmt.Errorf("get default style: %w", err)
	}
	return v, nil
}

func (s *PostgresStore) SaveDefault(ctx context.Context, v Style) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO default_message_style
		(singleton, font_family, font_size, color, bold, italic, underline)
		VALUES (TRUE, $1, $2, $3, $4, $5, $6)
		ON CONFLICT (singleton) DO UPDATE SET
		font_family = EXCLUDED.font_family, font_size = EXCLUDED.font_size,
		color = EXCLUDED.color, bold = EXCLUDED.bold, italic = EXCLUDED.italic,
		underline = EXCLUDED.underline, updated_at = now()`,
		v.FontFamily, v.FontSize, v.Color, v.Bold, v.Italic, v.Underline,
	)
	if err != nil {
		return fmt.Errorf("save default style: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetRoom(ctx context.Context, id string) (Style, bool, error) {
	var v Style
	err := s.db.QueryRowContext(ctx, `SELECT font_family, font_size, color, bold, italic, underline
		FROM room_message_styles WHERE room_id = $1`, id).Scan(
		&v.FontFamily, &v.FontSize, &v.Color, &v.Bold, &v.Italic, &v.Underline,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Style{}, false, nil
	}
	if err != nil {
		return Style{}, false, fmt.Errorf("get room style: %w", err)
	}
	return v, true, nil
}

func (s *PostgresStore) SaveRoom(ctx context.Context, id string, v Style) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO room_message_styles
		(room_id, font_family, font_size, color, bold, italic, underline)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (room_id) DO UPDATE SET
		font_family = EXCLUDED.font_family, font_size = EXCLUDED.font_size,
		color = EXCLUDED.color, bold = EXCLUDED.bold, italic = EXCLUDED.italic,
		underline = EXCLUDED.underline, updated_at = now()`,
		id, v.FontFamily, v.FontSize, v.Color, v.Bold, v.Italic, v.Underline,
	)
	if err != nil {
		return fmt.Errorf("save room style: %w", err)
	}
	return nil
}

func (s *PostgresStore) DeleteRoom(ctx context.Context, id string) error {
	if _, err := s.db.ExecContext(ctx, "DELETE FROM room_message_styles WHERE room_id = $1", id); err != nil {
		return fmt.Errorf("delete room style: %w", err)
	}
	return nil
}
