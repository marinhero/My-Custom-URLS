# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Run

```bash
# Run the application
go run .

# Build binary
go build -o shortme

# Run as API server
./shortme --server

# Run TUI against remote API
API_URL=https://your-app.onrender.com ./shortme
```

## Testing

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

Tests use in-memory SQLite, so no database setup is required.

## Database Configuration

Configuration is done via environment variables (see `.env.example`):
- `DB_ENGINE`: `postgres` (default), `mysql`, or `sqlite`
- `DB_USERNAME`, `DB_PASSWORD`, `DB_NAME`, `DB_HOST`, `DB_PORT`
- `REDIRECT_PREFIX_URL`
- `API_URL`: Remote API URL (enables remote TUI mode)

For quick local development with SQLite:
```bash
export DB_ENGINE=sqlite
export DB_NAME=shorturl.db
go run .
```

## Architecture

This is a URL shortener with two components:
1. **TUI Client** - Terminal interface built with Charm libraries
2. **REST API Server** - HTTP API for the TUI and external clients

```
┌─────────────────┐         ┌─────────────────┐
│  TUI Client     │  HTTP   │  API Server     │
│  (./shortme)    │ ──────► │  (--server)     │
└─────────────────┘         │  + Database     │
                            └─────────────────┘
```

### Libraries
- **GORM v2** (`gorm.io/gorm`) for database ORM
- **Bubble Tea** (`github.com/charmbracelet/bubbletea`) for TUI framework
- **Lipgloss** (`github.com/charmbracelet/lipgloss`) for styling
- **Bubbles** (`github.com/charmbracelet/bubbles`) for UI components

### Key Files

- `main.go` - Entry point, supports `--server` flag for API mode
- `model.go` - Bubble Tea model (Init, Update, View)
- `api.go` - REST API handlers
- `client.go` - HTTP client for remote API access
- `provider.go` - Interface for local/remote data access
- `db.go` - Database operations and models
- `commands.go` - Async commands (clipboard)
- `styles.go` - Lipgloss style definitions
- `config.go` - Environment-based configuration
- `db_test.go` - Test suite

### Deployment Files
- `Dockerfile` - Multi-stage build for container deployment
- `render.yaml` - Render.com Blueprint configuration

### Data Model

`CustomURL` struct: stores original URL, generated short URL, visit count, and timestamps.

### Key Types and Functions

- `URLProvider` interface: abstraction for local DB or remote API access
- `App` struct: implements URLProvider with direct DB access
- `APIClient` struct: implements URLProvider with HTTP client
- `model` struct: Bubble Tea model with state, table, input, URLs
- `API` struct: HTTP handlers for REST endpoints
- `validURL()`: validates URLs using `net/url` package
- `generateShortURL()`: creates random 10-character alphanumeric string

### API Endpoints

- `GET /health` - Health check
- `GET /api/urls` - List all URLs
- `POST /api/urls` - Create short URL
- `DELETE /api/urls/{id}` - Delete URL
- `GET /api/search?q=` - Search URLs
- `GET /r/{code}` - Redirect to original URL
