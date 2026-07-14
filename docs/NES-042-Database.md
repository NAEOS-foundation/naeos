# NES-042: Database Layer

## Overview

NAEOS provides a pluggable database layer supporting PostgreSQL, MySQL, and SQLite with production-ready features.

## Architecture

```
┌─────────────────────────────────────────┐
│              Application                │
├─────────────────────────────────────────┤
│         Database Interface              │
│  ┌─────────┬─────────┬─────────────┐   │
│  │PostgreSQL│ MySQL  │   SQLite    │   │
│  └─────────┴─────────┴─────────────┘   │
├─────────────────────────────────────────┤
│           Features                      │
│  • Context support (timeouts/cancel)   │
│  • Connection pooling                  │
│  • Retry with exponential backoff      │
│  • Query logging (slow query detect)   │
│  • Health checks                       │
│  • File-based migrations               │
└─────────────────────────────────────────┘
```

## Configuration

```go
type Config struct {
    Host            string
    Port            int
    User            string
    Password        string
    Database        string
    SSLMode         string        // disable, require, verify-ca, verify-full
    Timeout         time.Duration
    MaxOpenConns    int           // default: 25
    MaxIdleConns    int           // default: 5
    ConnMaxLifetime time.Duration
    ConnMaxIdleTime time.Duration
}
```

## Usage

### Factory Pattern

```go
// Create by driver name
db := database.New("postgresql")

// Create with config validation
db, err := database.NewFromConfig("postgresql", &database.Config{
    Host:     "localhost",
    Port:     5432,
    User:     "admin",
    Password: "secret",
    Database: "naeos",
    SSLMode:  "disable",
})
```

### Context Support

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

result, err := db.ExecContext(ctx, "INSERT INTO users (name) VALUES (?)", "test")
rows, err := db.QueryContext(ctx, "SELECT * FROM users")
tx, err := db.BeginTx(ctx)
```

### Migrations

```go
migrations := []database.Migration{
    {
        Version: 1,
        Name:    "create_users",
        Up:      "CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT)",
        Down:    "DROP TABLE users",
    },
}

err := db.Migrate(migrations)
err = db.Rollback(0) // rollback all migrations
```

### File-Based Migrations

```go
migrations, err := database.LoadMigrations("./migrations")
// Files: 000001_create_users.up.sql, 000001_create_users.down.sql
```

### Retry Logic

```go
err := database.WithRetry(ctx, 3, 100*time.Millisecond, func(ctx context.Context) error {
    return db.Ping()
})
```

### Query Logging

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
db := database.NewLoggingDatabase(realDB, logger)
```

### Health Checks

```go
if err := db.HealthCheck(); err != nil {
    log.Printf("database unhealthy: %v", err)
}
```

## Connection Pool

| Setting | Default | Description |
|---------|---------|-------------|
| MaxOpenConns | 25 | Maximum open connections |
| MaxIdleConns | 5 | Maximum idle connections |
| ConnMaxLifetime | 5m | Maximum connection lifetime |
| ConnMaxIdleTime | 0 | Maximum idle connection time |

## CLI Commands

```bash
# Connect to database
naeos db connect --type postgresql --name mydb --host localhost --port 5432

# List connections
naeos db list

# Run migrations
naeos db migrate --name mydb

# Disconnect
naeos db disconnect --name mydb
```

## Build Tags

Use `nosql` build tag to exclude database drivers:

```bash
go build -tags nosql ./...
```
