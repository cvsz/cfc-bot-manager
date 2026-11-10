package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain sets up the test database
func TestMain(m *testing.M) {
	// Check for test database URL
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		fmt.Println("TEST_DATABASE_URL not set, skipping database tests")
		os.Exit(0)
	}

	// Run tests
	os.Exit(m.Run())
}

func setupTestDB(t *testing.T) (*sql.DB, *pgxpool.Pool, func()) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)

	pool, err := pgxpool.New(context.Background(), dsn)
	require.NoError(t, err)

	// Clean up function
	cleanup := func() {
		// Clean up test data
		db.Exec("TRUNCATE TABLE IF EXISTS bot_connections, api_keys, audit_logs, plugin_configurations, plugins, bots, rooms, sessions, memberships, role_permissions, permissions, roles, users, organizations CASCADE")
		db.Close()
		pool.Close()
	}

	return db, pool, cleanup
}

func TestMigrationUpDown(t *testing.T) {
	_, pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Run migrations up
	err := RunMigrationsContext(ctx, pool, "migrations")
	require.NoError(t, err)

	// Verify all tables exist
	tables := []string{
		"organizations", "users", "roles", "permissions", "role_permissions",
		"memberships", "sessions", "rooms", "bots", "plugins",
		"plugin_configurations", "audit_logs", "api_keys", "bot_connections",
	}

	for _, table := range tables {
		var exists bool
		err := pool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)", table).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "table %s should exist", table)
	}

	// Check migration version
	version, err := MigrationStatusContext(ctx, pool, "migrations")
	require.NoError(t, err)
	assert.Equal(t, int64(1), version)

	// Rollback migration
	err = MigrationDownContext(ctx, pool, "migrations")
	require.NoError(t, err)

	// Verify tables are gone
	for _, table := range tables {
		var exists bool
		err := pool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)", table).Scan(&exists)
		require.NoError(t, err)
		assert.False(t, exists, "table %s should not exist after rollback", table)
	}

	// Check migration version is 0
	version, err = MigrationStatusContext(ctx, pool, "migrations")
	require.NoError(t, err)
	assert.Equal(t, int64(0), version)
}

func TestMigrationRollbackTo(t *testing.T) {
	_, pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Run migrations up
	err := RunMigrationsContext(ctx, pool, "migrations")
	require.NoError(t, err)

	// Rollback to version 0
	err = MigrationDownToContext(ctx, pool, "migrations", 0)
	require.NoError(t, err)

	// Verify version is 0
	version, err := MigrationStatusContext(ctx, pool, "migrations")
	require.NoError(t, err)
	assert.Equal(t, int64(0), version)
}

func TestTransactionRollback(t *testing.T) {
	_, pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Run migrations first
	err := RunMigrationsContext(ctx, pool, "migrations")
	require.NoError(t, err)

	tm := NewTransactionManager(pool)

	// Test that transaction rollback works
	err = tm.WithTransaction(ctx, func(tx pgx.Tx) error {
		// Insert an organization
		orgID := uuid.New()
		_, err := tx.Exec(ctx, `
			INSERT INTO organizations (id, name, slug, description, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, orgID, "Test Org", "test-org", "Test description", time.Now(), time.Now())
		require.NoError(t, err)

		// Force rollback by returning error
		return fmt.Errorf("intentional rollback")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "intentional rollback")

	// Verify organization was not persisted
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM organizations WHERE slug = 'test-org'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestTransactionCommit(t *testing.T) {
	_, pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	err := RunMigrationsContext(ctx, pool, "migrations")
	require.NoError(t, err)

	tm := NewTransactionManager(pool)

	// Test that transaction commit works
	var orgID uuid.UUID
	err = tm.WithTransaction(ctx, func(tx pgx.Tx) error {
		orgID = uuid.New()
		_, err := tx.Exec(ctx, `
			INSERT INTO organizations (id, name, slug, description, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, orgID, "Test Org", "test-org-commit", "Test description", time.Now(), time.Now())
		return err
	})
	require.NoError(t, err)

	// Verify organization was persisted
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM organizations WHERE slug = 'test-org-commit'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestSavepointRollback(t *testing.T) {
	_, pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	err := RunMigrationsContext(ctx, pool, "migrations")
	require.NoError(t, err)

	tm := NewTransactionManager(pool)

	// Test savepoint rollback
	err = tm.WithTransaction(ctx, func(tx pgx.Tx) error {
		// Create first organization
		_, err := tx.Exec(ctx, `
			INSERT INTO organizations (id, name, slug, description, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, uuid.New(), "Org 1", "org-1", "First", time.Now(), time.Now())
		require.NoError(t, err)

		// Create savepoint
		_, err = tx.Exec(ctx, "SAVEPOINT sp1")
		require.NoError(t, err)

		// Create second organization
		_, err = tx.Exec(ctx, `
			INSERT INTO organizations (id, name, slug, description, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, uuid.New(), "Org 2", "org-2", "Second", time.Now(), time.Now())
		require.NoError(t, err)

		// Rollback to savepoint
		_, err = tx.Exec(ctx, "ROLLBACK TO SAVEPOINT sp1")
		require.NoError(t, err)

		// Release savepoint
		_, err = tx.Exec(ctx, "RELEASE SAVEPOINT sp1")
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	// Verify only first organization exists
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM organizations").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestTransactionIsolation(t *testing.T) {
	_, pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	err := RunMigrationsContext(ctx, pool, "migrations")
	require.NoError(t, err)

	tm := NewTransactionManager(pool)

	// Test READ COMMITTED isolation
	err = tm.WithTransactionIsolation(ctx, IsolationLevelReadCommitted, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "SELECT 1")
		return err
	})
	require.NoError(t, err)

	// Test REPEATABLE READ isolation
	err = tm.WithTransactionIsolation(ctx, IsolationLevelRepeatableRead, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "SELECT 1")
		return err
	})
	require.NoError(t, err)

	// Test SERIALIZABLE isolation
	err = tm.WithTransactionIsolation(ctx, IsolationLevelSerializable, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "SELECT 1")
		return err
	})
	require.NoError(t, err)
}

func TestMigrationTestHelper(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mh := NewMigrationTestHelper(db, "migrations")

	// Apply migrations
	err := mh.ApplyMigrations()
	require.NoError(t, err)

	// Verify version
	version, err := mh.GetVersion()
	require.NoError(t, err)
	assert.Equal(t, int64(1), version)

	// Verify tables exist
	expectedTables := []string{
		"organizations", "users", "roles", "permissions", "role_permissions",
		"memberships", "sessions", "rooms", "bots", "plugins",
		"plugin_configurations", "audit_logs", "api_keys", "bot_connections",
	}
	err = mh.VerifySchema(expectedTables)
	require.NoError(t, err)

	// Rollback
	err = mh.RollbackLast()
	require.NoError(t, err)

	version, err = mh.GetVersion()
	require.NoError(t, err)
	assert.Equal(t, int64(0), version)

	// Re-apply
	err = mh.ApplyMigrations()
	require.NoError(t, err)

	version, err = mh.GetVersion()
	require.NoError(t, err)
	assert.Equal(t, int64(1), version)
}

func TestReset(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mh := NewMigrationTestHelper(db, "migrations")

	// Apply migrations
	err := mh.ApplyMigrations()
	require.NoError(t, err)

	version, err := mh.GetVersion()
	require.NoError(t, err)
	assert.Equal(t, int64(1), version)

	// Reset (rollback and reapply)
	err = mh.Reset()
	require.NoError(t, err)

	version, err = mh.GetVersion()
	require.NoError(t, err)
	assert.Equal(t, int64(1), version)
}

func TestTransactionConcurrency(t *testing.T) {
	_, pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	err := RunMigrationsContext(ctx, pool, "migrations")
	require.NoError(t, err)

	tm := NewTransactionManager(pool)

	// Test concurrent transactions
	done := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			err := tm.WithTransaction(ctx, func(tx pgx.Tx) error {
				orgID := uuid.New()
				_, err := tx.Exec(ctx, `
					INSERT INTO organizations (id, name, slug, description, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6)
				`, orgID, fmt.Sprintf("Org %d", i), fmt.Sprintf("org-%d", i), fmt.Sprintf("Description %d", i), time.Now(), time.Now())
				return err
			})
			done <- err
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		err := <-done
		require.NoError(t, err)
	}

	// Verify all organizations were created
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM organizations WHERE slug LIKE 'org-%'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 10, count)
}