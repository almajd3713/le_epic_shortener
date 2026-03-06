# Backend

The backend is a Go HTTP server that shortens URLs and redirects short codes to their original destinations.

## Package layout

```
backend/
├── cmd/server/main.go          Entry point — wires dependencies and starts Gin
├── internal/
│   ├── db/
│   │   ├── postgres.go         Creates the pgxpool connection pool
│   │   ├── migrate.go          Runs embedded SQL migrations at startup
│   │   └── migrations/
│   │       └── 001_create_urls.sql
│   ├── models/url.go           URL struct + request/response types
│   ├── repository/
│   │   └── url.repository.go  Raw SQL queries (Create, GetByCode, Deactivate)
│   ├── services/
│   │   ├── shortener.service.go   Generates short codes, calls repository
│   │   └── redirector.service.go  Looks up a code, returns the original URL
│   ├── handlers/
│   │   ├── shorten.go          POST /api/shorten
│   │   └── redirect.go         GET /:code
│   ├── middleware/
│   │   └── logger.go           Per-request structured logging + request ID
│   └── server/
│       └── routes.go           Registers all routes on the Gin engine
```

## Startup sequence

1. `godotenv` loads `.env` into the environment.
2. A `slog.Logger` is configured from `LOG_LEVEL` and `ENV` env vars (text in dev, JSON in production).
3. A `pgxpool` connection pool is created from `DATABASE_URL`.
4. `RunMigrations` applies any unapplied `.sql` files from `internal/db/migrations/`.
5. Dependencies flow down: `pool → repository → services → handlers`.
6. `SetupRoutes` registers the three routes on the Gin engine.
7. The server listens on `PORT` (default `8080`).

## Request lifecycle

### Shorten (`POST /api/shorten`)

```
Client → ShortenerHandler.POST
           └─ validates JSON body (long_url required, expires_at optional)
           └─ ShortenerService.ShortenURL(longUrl)
                └─ generates an 8-char nanoid
                └─ checks uniqueness via URLRepository.GetByCode
                └─ retries if the code already exists
                └─ URLRepository.Create → INSERT INTO urls
           └─ returns {"shortened_url": "<code>"}
```

### Redirect (`GET /:code`)

```
Client → RedirectHandler.GET
           └─ reads :code path param
           └─ RedirectorService.Redirect(code)
                └─ URLRepository.GetByCode
                     └─ SELECT WHERE short_code = $1 AND is_active = TRUE
                          AND (expires_at IS NULL OR expires_at > NOW())
           └─ 302 redirect to original URL  (404 if not found / expired)
```

## Database

A single table, `urls`, is created by `001_create_urls.sql`.

The partial index (`WHERE is_active = TRUE`) is what the redirect hot-path query hits.
As soft-deleted/expired rows accumulate they stay out of the index, keeping it small.

Migrations are tracked in a `schema_migrations` table (`filename TEXT PRIMARY KEY`).
`RunMigrations` reads all `*.sql` files embedded in the binary, skips already-applied ones,
and runs the rest in filename order.

## Middleware

`middleware.Logger` runs on every request and:

- Generates a UUID request ID, stored as `requestID` in the Gin context.
- Creates a child `slog.Logger` with `request_id`, `method`, and `path` fields, stored as `logger` in the context.
- After the handler returns, logs status code and duration at `INFO` level.

Handlers retrieve the logger with `c.Get("logger")` so log lines share the same request ID.

## Key dependencies

| Package | Purpose |
|---------|---------|
| `github.com/gin-gonic/gin` | HTTP router and middleware chain |
| `github.com/jackc/pgx/v5` | PostgreSQL driver + connection pool |
| `github.com/matoous/go-nanoid/v2` | URL-safe random ID generation |
| `github.com/google/uuid` | Request ID generation in the logger middleware |
| `github.com/joho/godotenv` | Loads `.env` in development |

## Environment variables

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | PostgreSQL connection string |
| `PORT` | No | Listen port (default `8080`) |
| `ENV` | No | Set to `production` for JSON-formatted logs |
| `LOG_LEVEL` | No | `DEBUG`, `INFO`, `WARN`, or `ERROR` (default `INFO`) |
