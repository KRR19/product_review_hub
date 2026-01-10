// Package testutil provides utilities for integration testing.
package testutil

import (
	"context"
	"fmt"
	"os"
	"product_review_hub/internal/database"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// TestDB provides a database connection for integration tests.
type TestDB struct {
	DB *sqlx.DB
}

// NewTestDB creates a new test database connection.
// It reads configuration from environment variables with fallback to defaults.
func NewTestDB(t *testing.T) *TestDB {
	t.Helper()

	cfg := database.Config{
		Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:     getEnvOrDefault("TEST_DB_PORT", "5433"),
		User:     getEnvOrDefault("TEST_DB_USER", "postgres"),
		Password: getEnvOrDefault("TEST_DB_PASSWORD", "postgres"),
		DBName:   getEnvOrDefault("TEST_DB_NAME", "product_review_hub_test"),
		SSLMode:  getEnvOrDefault("TEST_DB_SSLMODE", "disable"),
	}

	db, err := database.New(cfg)
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
	}

	return &TestDB{DB: db}
}

// Cleanup cleans up the test database tables.
func (tdb *TestDB) Cleanup(t *testing.T) {
	t.Helper()

	ctx := context.Background()

	// Delete in correct order due to foreign key constraints
	_, err := tdb.DB.ExecContext(ctx, "DELETE FROM reviews")
	require.NoError(t, err, "failed to clean up reviews")

	_, err = tdb.DB.ExecContext(ctx, "DELETE FROM products")
	require.NoError(t, err, "failed to clean up products")
}

// Close closes the database connection.
func (tdb *TestDB) Close(t *testing.T) {
	t.Helper()

	if err := tdb.DB.Close(); err != nil {
		t.Errorf("failed to close database: %v", err)
	}
}

// TruncateTables truncates all tables in the database.
func (tdb *TestDB) TruncateTables(t *testing.T) {
	t.Helper()

	ctx := context.Background()

	_, err := tdb.DB.ExecContext(ctx, "TRUNCATE TABLE reviews, products RESTART IDENTITY CASCADE")
	require.NoError(t, err, "failed to truncate tables")
}

// SetupTestDB creates a new test database connection and registers cleanup.
func SetupTestDB(t *testing.T) *TestDB {
	t.Helper()

	tdb := NewTestDB(t)

	t.Cleanup(func() {
		tdb.Cleanup(t)
		tdb.Close(t)
	})

	return tdb
}

// MustExec executes a query and fails the test if it errors.
func (tdb *TestDB) MustExec(t *testing.T, query string, args ...interface{}) {
	t.Helper()

	_, err := tdb.DB.Exec(query, args...)
	require.NoError(t, err, "failed to execute query %q", query)
}

// CreateTestProduct creates a test product and returns its ID.
func (tdb *TestDB) CreateTestProduct(t *testing.T, name string, description *string, price float64) int64 {
	t.Helper()

	var id int64
	query := `INSERT INTO products (name, description, price) VALUES ($1, $2, $3) RETURNING id`
	err := tdb.DB.QueryRowx(query, name, description, price).Scan(&id)
	require.NoError(t, err, "failed to create test product")

	return id
}

// CreateTestReview creates a test review and returns its ID.
func (tdb *TestDB) CreateTestReview(t *testing.T, productID int64, firstName, lastName string, rating int, comment *string) int64 {
	t.Helper()

	var id int64
	query := `INSERT INTO reviews (product_id, first_name, last_name, rating, comment) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := tdb.DB.QueryRowx(query, productID, firstName, lastName, rating, comment).Scan(&id)
	require.NoError(t, err, "failed to create test review")

	return id
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// StringPtr returns a pointer to the given string.
func StringPtr(s string) *string {
	return &s
}

// Float64Ptr returns a pointer to the given float64.
func Float64Ptr(f float64) *float64 {
	return &f
}

// IntPtr returns a pointer to the given int.
func IntPtr(i int) *int {
	return &i
}

// RequireIntegrationTest skips the test if not running integration tests.
func RequireIntegrationTest(t *testing.T) {
	t.Helper()

	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=true to run.")
	}
}

// TestDSN returns the DSN for the test database.
func TestDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnvOrDefault("TEST_DB_HOST", "localhost"),
		getEnvOrDefault("TEST_DB_PORT", "5433"),
		getEnvOrDefault("TEST_DB_USER", "postgres"),
		getEnvOrDefault("TEST_DB_PASSWORD", "postgres"),
		getEnvOrDefault("TEST_DB_NAME", "product_review_hub_test"),
		getEnvOrDefault("TEST_DB_SSLMODE", "disable"),
	)
}
