# Product Review Hub

Microservice system for managing products and reviews with support for caching, event-driven architecture, and high reliability.

## Project Description

Product Review Hub is a RESTful API service for managing a product catalog and user reviews. The system provides full CRUD for products and reviews, automatic calculation of product average ratings, data caching, and notification of external services about events via a message broker.

### Key Features

- **Product Management**: create, edit, delete, retrieve list and individual product
- **Review Management**: create, edit, delete reviews for products
- **Automatic Average Rating Calculation** for products based on all reviews
- **Caching** of reviews and ratings in Redis to improve performance
- **Event-Driven Model**: notification of external services upon review creation/modification/deletion via RabbitMQ
- **Idempotency Mechanism** for safe retry requests
- **OpenAPI Specification** with automatic code generation

### Data Structure

**Product**
- `id` — unique identifier
- `name` — product name
- `description` — description
- `price` — price
- `average_rating` — average rating (computed field)

**Review**
- `id` — unique identifier
- `product_id` — product ID
- `first_name` — author's first name
- `last_name` — author's last name
- `rating` — rating (1-5)
- `comment` — review text

## Technology Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.22+ |
| HTTP Router | Chi v5 |
| Database | PostgreSQL 16 |
| Cache | Redis 7 |
| Message Broker | RabbitMQ 3.12 |
| Migrations | golang-migrate |
| API Specification | OpenAPI 3.0 |
| Code Generation | oapi-codegen |

## Running via Docker Compose

### Requirements

- Docker 20.10+
- Docker Compose 2.0+

### Quick Start

```bash
# Clone the repository
git clone <repository-url>
cd product_review_hub

# Start all services
docker-compose up --build
```

Or using Makefile:

```bash
make run
```

After startup, the following will be available:

| Service | URL |
|--------|-----|
| API | http://localhost:8080 |
| RabbitMQ Management UI | http://localhost:15672 (guest/guest) |
| PostgreSQL | localhost:5432 |
| Redis | localhost:6379 |

### Health Check

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

### Container Management

```bash
# Run in background
make docker-up

# Stop
make docker-down

# View logs
make docker-logs
```

## API Endpoints

### Products

| Method | Endpoint | Description |
|-------|----------|----------|
| `POST` | `/api/v1/products` | Create product |
| `GET` | `/api/v1/products` | Get list of products |
| `GET` | `/api/v1/products/{id}` | Get product by ID |
| `PUT` | `/api/v1/products/{id}` | Update product |
| `DELETE` | `/api/v1/products/{id}` | Delete product |

### Reviews

| Method | Endpoint | Description |
|-------|----------|----------|
| `POST` | `/api/v1/products/{productId}/reviews` | Create review |
| `GET` | `/api/v1/products/{productId}/reviews` | Get product reviews |
| `PUT` | `/api/v1/products/{productId}/reviews/{reviewId}` | Update review |
| `DELETE` | `/api/v1/products/{productId}/reviews/{reviewId}` | Delete review |

### Request Examples

```bash
# Create product
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{"name": "iPhone 15", "description": "Apple Smartphone", "price": 999.99}'

# Create review
curl -X POST http://localhost:8080/api/v1/products/1/reviews \
  -H "Content-Type: application/json" \
  -d '{"first_name": "Ivan", "last_name": "Petrov", "rating": 5, "comment": "Excellent product!"}'

# Get product with rating
curl http://localhost:8080/api/v1/products/1
```

### Idempotency

For safe retry requests, use the `X-Idempotency-Key` header:

```bash
curl -X POST http://localhost:8080/api/v1/products/1/reviews \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: unique-request-id-123" \
  -d '{"rating": 5, "comment": "Excellent!"}'
```

## Running Tests

### Requirements

- Go 1.22+
- Docker and Docker Compose (for E2E tests)

### Unit Tests

```bash
make test
```

### E2E Tests

E2E tests automatically bring up a test database:

```bash
make test-e2e
```

### All Tests with Test Database

```bash
make test-with-docker
```

### Code Coverage

```bash
make test-coverage
```

After execution, a `coverage.html` file with a visual report will be created.

## Architectural Decisions and Trade-offs

### 1. Layered Architecture

The project is built on clean architecture principles with separation into layers:

```
cmd/                    # Application entry points
├── api/               # HTTP API server
└── review-watcher/    # Event consumer service

internal/
├── api/               # Generated code from OpenAPI
├── cache/             # Caching service
├── config/            # Application configuration
├── database/          # DB initialization
├── handler/           # HTTP handlers
├── middleware/        # Middleware (idempotency)
├── models/            # Domain models
├── rabbitmq/          # RabbitMQ operations
├── redis/             # Redis client
├── repository/        # Data access layer
└── server/            # HTTP server
```

**Trade-off**: Layered architecture adds boilerplate code but ensures testability and component replaceability.

### 2. Interfaces for Repositories

Repositories are defined via interfaces in the handlers layer:

```go
type ProductRepository interface {
    Create(ctx context.Context, tx *sqlx.Tx, params models.CreateProductParams) (*models.Product, error)
    GetByID(ctx context.Context, tx *sqlx.Tx, id int64) (*models.ProductWithRating, error)
    // ...
}
```

**Advantages**:
- Simplicity of unit testing via mock objects
- Independence of business logic from storage implementation

**Trade-off**: Additional abstractions complicate the code but are critical for testability.

### 3. Transactions in Repositories

Each operation runs in the context of a transaction:

```go
tx, err := h.ProductRepo.BeginTx(r.Context())
defer tx.Rollback()
// ... operations ...
h.ProductRepo.CommitTx(r.Context(), tx)
```

**Advantages**:
- Atomicity of complex operations
- Consistency when working with multiple tables

**Trade-off**: Increased load on DB connections, but guarantees data integrity.

### 4. Caching (Cache-Aside Pattern)

The Cache-Aside pattern is used for caching reviews and ratings:

```go
// Attempt to get from cache
cachedReviews, err := h.Cache.GetReviews(ctx, productID, limit, offset)
if cachedReviews != nil {
    return cachedReviews // Cache hit
}

// Cache miss - load from DB
reviews := h.ReviewRepo.ListByProductID(...)

// Save to cache
h.Cache.SetReviews(ctx, productID, limit, offset, reviews)
```

**Cache Invalidation** occurs on any review change:

```go
h.Cache.InvalidateProductCache(ctx, productID)
```

**Trade-off**:
- Cache TTL (5 minutes) — compromise between freshness and DB load
- Invalidation of all product keys on change — simpler implementation, but less efficient than targeted invalidation

### 5. Event-Driven Architecture (Event-Driven)

Events are published to RabbitMQ upon review changes:

```go
event := rabbitmq.NewReviewEvent(
    rabbitmq.EventReviewCreated,
    reviewID,
    productID,
    rating,
)
h.Publisher.Publish(ctx, event)
```

Event types:
- `review.created`
- `review.updated`
- `review.deleted`

**Review Watcher** — a demonstration service subscribed to events:

```
=== Review Event Received ===
  Event Type: review.created
  Review ID:  1
  Product ID: 1
  Rating:     5
=============================
```

**Trade-off**:
- Asynchronous publication without delivery guarantee (at-most-once) — transactional outbox is required for critical systems
- Topic exchange with wildcard routing — flexible, but requires naming conventions

### 6. Idempotency Middleware

Idempotency mechanism to protect against duplicate requests:

```go
func Idempotency(store idempotency.Store, ttl time.Duration) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        // Check cache by X-Idempotency-Key
        // Cache successful responses
    }
}
```

**Trade-off**:
- Storage in Redis with 1-minute TTL — balance between protection and memory
- Only 2xx responses are cached — errors can be retried

### 7. Average Rating Calculation

Average rating is calculated on each product request via JOIN:

```sql
SELECT p.*, AVG(r.rating)::FLOAT AS average_rating
FROM products p
LEFT JOIN reviews r ON p.id = r.product_id
WHERE p.id = $1
GROUP BY p.id
```

**Trade-off**:
- Simplicity of implementation vs denormalization
- For high-load systems, a materialized view or pre-computed column is recommended

### 8. OpenAPI First Approach

API is described in OpenAPI 3.0, code is generated automatically:

```bash
make generate  # oapi-codegen generates internal/api/generated.go
```

**Advantages**:
- Single source of truth for API contract
- Automatic validation and typing

**Trade-off**: Dependency on code generator, limited customization flexibility.

### 9. Graceful Shutdown

Services correctly terminate operations on signals:

```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
<-quit
srv.Shutdown(ctx)
```

**Advantages**: Correct closure of connections to DB, Redis, RabbitMQ.

### 10. Configuration via Environment Variables

12-factor app approach:

```go
cfg := &Config{
    ServerAddress: getEnv("SERVER_ADDRESS", ":8080"),
    Database: database.Config{
        Host: getEnv("DB_HOST", "localhost"),
        // ...
    },
}
```

**Advantages**: Simplicity of deployment in various environments.

## Possible Improvements

1. **Structured logging** — replace standard logger with zap/zerolog
2. **Metrics** — Prometheus metrics for monitoring
3. **Tracing** — OpenTelemetry for distributed tracing
4. **Rate limiting** — overload protection
5. **Circuit breaker** — for external dependencies
6. **Transactional outbox** — guaranteed event delivery
7. **Materialized view** — for average rating under high load

## Project Structure

```
.
├── api/                    # OpenAPI specification
│   ├── openapi.yaml
│   └── oapi-codegen.yaml
├── cmd/
│   ├── api/               # Main API server
│   └── review-watcher/    # Event service
├── internal/
│   ├── api/               # Generated code
│   ├── cache/             # Redis caching
│   ├── config/            # Configuration
│   ├── database/          # DB connection
│   ├── handler/           # HTTP handlers
│   ├── middleware/        # Middleware
│   ├── models/            # Domain models
│   ├── rabbitmq/          # Queue operations
│   ├── redis/             # Redis client
│   ├── repository/        # Repositories
│   └── server/            # HTTP server
├── migrations/            # SQL migrations
├── tests/
│   └── e2e/              # End-to-end tests
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── README.md
```
