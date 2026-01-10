// Package e2e provides end-to-end tests infrastructure.
package e2e

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"product_review_hub/internal/api"
	"product_review_hub/internal/database"
	"product_review_hub/internal/handler"
	"product_review_hub/internal/repository/products"
	"product_review_hub/internal/repository/reviews"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// TestEnv holds the test environment configuration and resources.
type TestEnv struct {
	DB      *sqlx.DB
	Server  *httptest.Server
	Client  *http.Client
	BaseURL string
}

// Setup creates a new test environment with database connection and HTTP server.
func Setup(t *testing.T) *TestEnv {
	t.Helper()

	db := setupDatabase(t)
	server := setupServer(t, db)

	return &TestEnv{
		DB:      db,
		Server:  server,
		Client:  &http.Client{Timeout: 10 * time.Second},
		BaseURL: server.URL,
	}
}

// Teardown cleans up test environment resources.
func (env *TestEnv) Teardown(t *testing.T) {
	t.Helper()

	if env.Server != nil {
		env.Server.Close()
	}

	if env.DB != nil {
		env.DB.Close()
	}
}

// CleanupProducts removes all products from the database.
func (env *TestEnv) CleanupProducts(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := env.DB.ExecContext(ctx, "TRUNCATE TABLE products CASCADE")
	require.NoError(t, err, "Failed to cleanup products")
}

// CleanupReviews removes all reviews from the database.
func (env *TestEnv) CleanupReviews(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := env.DB.ExecContext(ctx, "TRUNCATE TABLE reviews CASCADE")
	require.NoError(t, err, "Failed to cleanup reviews")
}

// setupDatabase creates a database connection using test environment variables.
func setupDatabase(t *testing.T) *sqlx.DB {
	t.Helper()

	cfg := database.Config{
		Host:            getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:            getEnvOrDefault("TEST_DB_PORT", "5433"),
		User:            getEnvOrDefault("TEST_DB_USER", "postgres"),
		Password:        getEnvOrDefault("TEST_DB_PASSWORD", "postgres"),
		DBName:          getEnvOrDefault("TEST_DB_NAME", "product_review_hub_test"),
		SSLMode:         getEnvOrDefault("TEST_DB_SSLMODE", "disable"),
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Minute,
	}

	db, err := database.New(cfg)
	require.NoError(t, err, "Failed to connect to test database")

	return db
}

// setupServer creates a test HTTP server with all routes.
func setupServer(t *testing.T, db *sqlx.DB) *httptest.Server {
	t.Helper()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	productRepo := products.NewRepository(db)
	reviewRepo := reviews.NewRepository(db)
	h := handler.New(db, productRepo, reviewRepo, nil) // nil publisher for tests

	api.HandlerFromMux(h, r)

	return httptest.NewServer(r)
}

// getEnvOrDefault returns the value of an environment variable or a default value.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// APIEndpoint returns the full URL for an API endpoint.
func (env *TestEnv) APIEndpoint(path string) string {
	return fmt.Sprintf("%s%s", env.BaseURL, path)
}
