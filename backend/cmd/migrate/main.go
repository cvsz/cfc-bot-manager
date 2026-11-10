package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/cvsz/cfc-bot-manager/backend/internal/messagestyle"
	"github.com/cvsz/cfc-bot-manager/backend/internal/persistence"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("database unavailable: %v", err)
	}

	// Run core schema migrations (from migrations/ directory)
	coreMigCfg := persistence.DefaultConfig()
	coreMigCfg.MigrationPath = "migrations"
	if err := runGooseMigrations(ctx, db, coreMigCfg.MigrationPath); err != nil {
		log.Fatalf("apply core migrations: %v", err)
	}
	log.Print("core schema migrations applied")

	// Run message style migrations (embedded)
	if err := messagestyle.ApplyMigrations(ctx, db); err != nil {
		log.Fatalf("apply message style migrations: %v", err)
	}
	log.Print("message style migrations applied")

	log.Print("all database migrations applied successfully")
}

func runGooseMigrations(ctx context.Context, db *sql.DB, migrationPath string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	// Set base FS to nil to use filesystem
	goose.SetBaseFS(nil)

	if err := goose.UpContext(ctx, db, migrationPath); err != nil {
		return err
	}
	return nil
}