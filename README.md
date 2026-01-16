# ShortMe - URL Shortener

A URL shortener with a beautiful TUI client and REST API server. Deploy the API to the cloud and use the TUI from anywhere.

## Features

- **TUI Client**: Beautiful terminal interface built with Charm libraries
- **REST API**: JSON API for programmatic access
- **Cloud Ready**: Deploy to Render.com with one click
- Create, list, delete, and search short URLs
- Copy short URLs to clipboard
- Visit tracking
- Supports PostgreSQL, MySQL, and SQLite

## Architecture

```
┌─────────────────┐         ┌─────────────────┐
│  TUI Client     │  HTTP   │  API Server     │
│  (local)        │ ──────► │  (Render.com)   │
└─────────────────┘         │  + PostgreSQL   │
                            └─────────────────┘
```

## Installation

```bash
git clone <repository>
cd my-custom-urls
go build -o shortme
```

## Running Modes

### 1. Local Mode (TUI + Local Database)

```bash
export DB_ENGINE=sqlite
export DB_NAME=shorturl.db
./shortme
```

### 2. Server Mode (API Server)

```bash
export DB_ENGINE=sqlite
export DB_NAME=shorturl.db
./shortme --server
```

### 3. Remote Mode (TUI → Remote API)

```bash
export API_URL=https://your-app.onrender.com
./shortme
```

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `API_URL` | Remote API URL (enables remote mode) | - |
| `DB_ENGINE` | Database: postgres, mysql, sqlite | postgres |
| `DB_USERNAME` | Database username | marin |
| `DB_PASSWORD` | Database password | devel |
| `DB_NAME` | Database name (or file for SQLite) | shorturl |
| `DB_HOST` | Database host | localhost |
| `DB_PORT` | Database port | 5432 |
| `SERVER_PORT` | API server port | 8080 |
| `REDIRECT_PREFIX_URL` | Base URL for short links | http://127.0.0.1:8080/r/ |

## Deploy to Render

[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy)

1. Fork this repository
2. Create a new Blueprint on Render
3. Connect your repo and deploy
4. Set `REDIRECT_PREFIX_URL` to `https://your-app.onrender.com/r/`

The `render.yaml` file configures:
- Web service running the API server
- PostgreSQL database

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/api/urls` | List all URLs |
| `POST` | `/api/urls` | Create short URL |
| `DELETE` | `/api/urls/{id}` | Delete URL |
| `GET` | `/api/search?q=` | Search URLs |
| `GET` | `/r/{code}` | Redirect to original URL |

### Example: Create Short URL

```bash
curl -X POST https://your-app.onrender.com/api/urls \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/very-long-url"}'
```

## TUI Key Bindings

| Key | Action |
|-----|--------|
| `n` | Create new short URL |
| `d` | Delete selected URL |
| `Enter` | Copy short URL to clipboard |
| `↑`/`↓` or `j`/`k` | Navigate list |
| `q` | Quit |
| `Esc` | Cancel current action |

## Development

```bash
# Run tests
go test -v ./...

# Run with coverage
go test -cover ./...

# Build
go build -o shortme

# Run API server locally
./shortme --server

# Run TUI in local mode
./shortme

# Run TUI against local server
API_URL=http://localhost:8080 ./shortme
```

## License

MIT
