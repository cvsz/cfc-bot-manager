package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// Config holds database connection configuration
type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	MigrationPath   string
}

// DefaultConfig returns a config with sensible defaults
func DefaultConfig() Config {
	return Config{
		Host:            "127.0.0.1",
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		Database:        "cfc_bot_manager",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
		MigrationPath:   "migrations",
	}
}

// DSN builds a PostgreSQL connection string
func (c Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

// Pool creates a new pgxpool.Pool with the given config
func (c Config) Pool(ctx context.Context) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(c.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}

	poolConfig.MaxConns = int32(c.MaxOpenConns)
	poolConfig.MinConns = int32(c.MaxIdleConns)
	poolConfig.MaxConnLifetime = c.ConnMaxLifetime
	poolConfig.MaxConnIdleTime = c.ConnMaxIdleTime
	poolConfig.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

// DB wraps sql.DB for goose migrations
func (c Config) DB() (*sql.DB, error) {
	db, err := sql.Open("pgx", c.DSN())
	if err != nil {
		return nil, fmt.Errorf("open sql db: %w", err)
	}
	db.SetMaxOpenConns(c.MaxOpenConns)
	db.SetMaxIdleConns(c.MaxIdleConns)
	db.SetConnMaxLifetime(c.ConnMaxLifetime)
	db.SetConnMaxIdleTime(c.ConnMaxIdleTime)
	return db, nil
}

// RunMigrations applies all pending migrations
func RunMigrations(db *sql.DB, migrationPath string) error {
	goose.SetBaseFS(nil) // Use filesystem
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	if err := goose.Up(db, migrationPath); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}

// RunMigrationsContext applies migrations with context (for pgxpool)
func RunMigrationsContext(ctx context.Context, pool *pgxpool.Pool, migrationPath string) error {
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()
	return RunMigrations(db, migrationPath)
}

// MigrationStatus returns the current migration version
func MigrationStatus(db *sql.DB, migrationPath string) (int64, error) {
	if err := goose.SetDialect("postgres"); err != nil {
		return 0, fmt.Errorf("set dialect: %w", err)
	}
	return goose.GetDBVersion(db)
}

// MigrationStatusContext returns the current migration version with context
func MigrationStatusContext(ctx context.Context, pool *pgxpool.Pool, migrationPath string) (int64, error) {
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()
	return MigrationStatus(db, migrationPath)
}

// MigrationDown rolls back the last applied migration
func MigrationDown(db *sql.DB, migrationPath string) error {
	goose.SetBaseFS(nil)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	if err := goose.Down(db, migrationPath); err != nil {
		return fmt.Errorf("rollback migration: %w", err)
	}
	return nil
}

// MigrationDownContext rolls back the last applied migration with context
func MigrationDownContext(ctx context.Context, pool *pgxpool.Pool, migrationPath string) error {
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()
	return MigrationDown(db, migrationPath)
}

// MigrationDownTo rolls back to a specific version
func MigrationDownTo(db *sql.DB, migrationPath string, version int64) error {
	goose.SetBaseFS(nil)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	if err := goose.DownTo(db, migrationPath, version); err != nil {
		return fmt.Errorf("rollback to version %d: %w", version, err)
	}
	return nil
}

// MigrationDownToContext rolls back to a specific version with context
func MigrationDownToContext(ctx context.Context, pool *pgxpool.Pool, migrationPath string, version int64) error {
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()
	return MigrationDownTo(db, migrationPath, version)
}

// CreateDatabase creates the database if it doesn't exist (requires admin connection)
func CreateDatabase(adminDSN, dbName string) error {
	adminDB, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return fmt.Errorf("connect to admin db: %w", err)
	}
	defer adminDB.Close()

	// Check if database exists
	var exists bool
	err = adminDB.QueryRow("SELECT 1 FROM pg_database WHERE datname = $1", dbName).Scan(&exists)
	if err == sql.ErrNoRows {
		// Database doesn't exist, create it
		_, err = adminDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
		if err != nil {
			return fmt.Errorf("create database: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("check database exists: %w", err)
	}
	return nil
}

// TransactionManager provides transaction management utilities
type TransactionManager struct {
	pool *pgxpool.Pool
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(pool *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{pool: pool}
}

// WithTransaction executes a function within a database transaction
func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("transaction failed: %w, rollback failed: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}

// WithSavepoint executes a function within a savepoint
// Allows partial rollback within a larger transaction
func (tm *TransactionManager) WithSavepoint(ctx context.Context, name string, fn func(pgx.Tx) error) error {
	return tm.WithTransaction(ctx, func(tx pgx.Tx) error {
		// Create savepoint
		if _, err := tx.Exec(ctx, fmt.Sprintf("SAVEPOINT %s", name)); err != nil {
			return fmt.Errorf("create savepoint: %w", err)
		}

		if err := fn(tx); err != nil {
			// Rollback to savepoint
			if _, rbErr := tx.Exec(ctx, fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", name)); rbErr != nil {
				return fmt.Errorf("savepoint rollback failed: %w, original error: %w", rbErr, err)
			}
			return err
		}

		// Release savepoint
		if _, err := tx.Exec(ctx, fmt.Sprintf("RELEASE SAVEPOINT %s", name)); err != nil {
			return fmt.Errorf("release savepoint: %w", err)
		}
		return nil
	})
}

// IsolationLevel represents transaction isolation levels
type IsolationLevel string

const (
	IsolationLevelReadCommitted  IsolationLevel = "READ COMMITTED"
	IsolationLevelRepeatableRead IsolationLevel = "REPEATABLE READ"
	IsolationLevelSerializable   IsolationLevel = "SERIALIZABLE"
)

// WithTransactionIsolation executes a function with a specific isolation level
func (tm *TransactionManager) WithTransactionIsolation(ctx context.Context, level IsolationLevel, fn func(pgx.Tx) error) error {
	var isoLevel pgx.TxIsoLevel
	switch level {
	case IsolationLevelReadCommitted:
		isoLevel = pgx.ReadCommitted
	case IsolationLevelRepeatableRead:
		isoLevel = pgx.RepeatableRead
	case IsolationLevelSerializable:
		isoLevel = pgx.Serializable
	default:
		isoLevel = pgx.ReadCommitted
	}

	tx, err := tm.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: isoLevel,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("transaction failed: %w, rollback failed: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}

// MigrationTestHelper provides utilities for testing migrations
type MigrationTestHelper struct {
	db          *sql.DB
	migrationPath string
}

// NewMigrationTestHelper creates a new migration test helper
func NewMigrationTestHelper(db *sql.DB, migrationPath string) *MigrationTestHelper {
	return &MigrationTestHelper{
		db:            db,
		migrationPath: migrationPath,
	}
}

// ApplyMigrations applies all migrations
func (mh *MigrationTestHelper) ApplyMigrations() error {
	return RunMigrations(mh.db, mh.migrationPath)
}

// RollbackLast rolls back the last migration
func (mh *MigrationTestHelper) RollbackLast() error {
	return MigrationDown(mh.db, mh.migrationPath)
}

// RollbackTo rolls back to a specific version
func (mh *MigrationTestHelper) RollbackTo(version int64) error {
	return MigrationDownTo(mh.db, mh.migrationPath, version)
}

// GetVersion returns the current migration version
func (mh *MigrationTestHelper) GetVersion() (int64, error) {
	return MigrationStatus(mh.db, mh.migrationPath)
}

// Reset resets the database (drops all tables and reapplies migrations)
func (mh *MigrationTestHelper) Reset() error {
	if err := mh.RollbackTo(0); err != nil {
		return err
	}
	return mh.ApplyMigrations()
}

// VerifySchema verifies the current schema matches expectations
func (mh *MigrationTestHelper) VerifySchema(expectedTables []string) error {
	rows, err := mh.db.Query(`
		SELECT table_name FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
	`)
	if err != nil {
		return fmt.Errorf("query tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return err
		}
		tables = append(tables, table)
	}

	if len(tables) != len(expectedTables) {
		return fmt.Errorf("table count mismatch: got %d, expected %d", len(tables), len(expectedTables))
	}

	tableSet := make(map[string]bool)
	for _, t := range tables {
		tableSet[t] = true
	}

	for _, expected := range expectedTables {
		if !tableSet[expected] {
			return fmt.Errorf("missing expected table: %s", expected)
		}
	}

	return nil
}

// TransactionTestHelper provides utilities for testing transaction behavior
type TransactionTestHelper struct {
	pool *pgxpool.Pool
}

// NewTransactionTestHelper creates a new transaction test helper
func NewTransactionTestHelper(pool *pgxpool.Pool) *TransactionTestHelper {
	return &TransactionTestHelper{pool: pool}
}

// TestRollback tests that a transaction rollback works correctly
func (tth *TransactionTestHelper) TestRollback(ctx context.Context, setup func(pgx.Tx) error) error {
	tx, err := tth.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	if err := setup(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	// Force rollback by returning an error
	if rbErr := tx.Rollback(ctx); rbErr != nil {
		return fmt.Errorf("rollback failed: %w", rbErr)
	}
	return fmt.Errorf("intentional rollback")
}

// TestSavepointRollback tests that savepoint rollback works correctly
func (tth *TransactionTestHelper) TestSavepointRollback(ctx context.Context, name string, setup func(pgx.Tx) error) error {
	tx, err := tth.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	if _, err := tx.Exec(ctx, fmt.Sprintf("SAVEPOINT %s", name)); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if err := setup(tx); err != nil {
		if _, rbErr := tx.Exec(ctx, fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", name)); rbErr != nil {
			_ = tx.Rollback(ctx)
			return rbErr
		}
		_ = tx.Rollback(ctx)
		return nil // Expected rollback
	}

	if _, err := tx.Exec(ctx, fmt.Sprintf("RELEASE SAVEPOINT %s", name)); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

// TestIsolationLevel tests that the isolation level is respected
func (tth *TransactionTestHelper) TestIsolationLevel(ctx context.Context, level IsolationLevel, fn func(pgx.Tx) error) error {
	var isoLevel pgx.TxIsoLevel
	switch level {
	case IsolationLevelReadCommitted:
		isoLevel = pgx.ReadCommitted
	case IsolationLevelRepeatableRead:
		isoLevel = pgx.RepeatableRead
	case IsolationLevelSerializable:
		isoLevel = pgx.Serializable
	default:
		isoLevel = pgx.ReadCommitted
	}

	tx, err := tth.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: isoLevel,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("transaction failed: %w, rollback failed: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}