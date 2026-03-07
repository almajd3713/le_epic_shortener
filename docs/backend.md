# Backend

The backend is a Go HTTP server that shortens URLs and redirects short codes to their original destinations.

## Package layout

```
backend/
├── cmd/server/main.go          Entry point — wires dependencies and starts Gin
├── internal/
│   ├── config/
│   │   └── config.go           Loads all env vars into a typed Config struct
│   ├── db/
│   │   ├── postgres.go         Creates the pgxpool connection pool
│   │   ├── migrate.go          Runs embedded SQL migrations at startup
│   │   └── migrations/
│   │       └── 001_create_urls.sql
│   ├── models/url.go           URL struct + request/response types (with JSON tags)
│   ├── repository/
│   │   └── url.repository.go   Raw SQL queries (Create, GetByCode, GetAll, Deactivate)
│   ├── services/
│   │   ├── url.service.go          GetAllURLs; defines IURLService interface
│   │   ├── shortener.service.go    Generates short codes, calls repository
│   │   └── redirector.service.go   Looks up a code, returns the original URL
│   ├── handlers/
│   │   ├── shorten.go          POST /api/shorten
│   │   ├── redirect.go         GET /:code
│   │   └── urls.go             GET /api/urls
│   ├── middleware/
│   │   ├── logger.go           Per-request structured logging + request ID
│   │   └── cors.go             CORS middleware wrapper around gin-contrib/cors
│   └── server/
│       └── routes.go           Registers all routes on the Gin engine
```

## Startup sequence

1. `godotenv` loads `.env` into the environment.
2. `config.Load()` reads all env vars into a `Config` struct — no env access beyond this point.
3. A `slog.Logger` is configured from `LOG_LEVEL` and `ENV` (text handler in dev, JSON in production).
4. A `pgxpool` connection pool is created from `DATABASE_URL`.
5. `RunMigrations` applies any unapplied `.sql` files embedded in the binary.
6. Dependencies flow down: `pool → repository → services → handlers`.
7. CORS is configured from `ALLOWED_ORIGINS` and applied as middleware.
8. `SetupRoutes` registers all routes on the Gin engine.
9. The server listens on `PORT`.

## Request lifecycle

### Shorten (`POST /api/shorten`)

```
Client → ShortenerHandler.POST
           └─ binds JSON body: { long_url, expires_at? }
           └─ ShortenerService.ShortenURL(longUrl)
                └─ generates an 8-char nanoid
                └─ checks uniqueness via URLRepository.GetByCode
                └─ retries if the code already exists (collision loop)
                └─ URLRepository.Create → INSERT INTO urls RETURNING *
           └─ returns { short_code, short_url, created_at }
                short_url = BASE_URL (from env via BaseURL middleware) + "/" + short_code
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

### List URLs (`GET /api/urls`)

```
Client → URLHandler.GET_ALL
           └─ URLService.GetAllURLs()
                └─ URLRepository.GetAll
                     └─ SELECT WHERE is_active = TRUE AND not expired
                          ORDER BY created_at DESC
           └─ returns JSON array of URL objects
```

## Models

`models.URL` is the core domain type shared across layers:

```go
type URL struct {
    ID        int64      `json:"id"`
    ShortCode string     `json:"short_code"`
    LongURL   string     `json:"long_url"`
    CreatedAt time.Time  `json:"created_at"`
    ExpiresAt *time.Time `json:"expires_at"` // nil = no expiration
    IsActive  bool       `json:"is_active"`
}
```

`models.URLResponse` is the HTTP response shape for `POST /api/shorten`:

```go
type URLResponse struct {
    ShortCode    string `json:"short_code"`
    ShortenedURL string `json:"shortened_url"` // full redirect URL: host + "/" + code
    CreatedAt    string `json:"created_at"`
}
```

## Service interface

`URLService` implements `IURLService`, which is the type `RedirectorService` depends on:

```go
type IURLService interface {
    GetOriginalURL(shortenedURL string) (string, error)
}
```

This keeps `RedirectorService` decoupled from the concrete `URLService` implementation — it only
needs the redirect lookup capability. Always pass the interface value (not a pointer to interface).

## Database

A single table, `urls`, is created by `001_create_urls.sql`.

The partial index (`WHERE is_active = TRUE`) is what the redirect hot-path query hits.
As soft-deleted/expired rows accumulate they stay out of the index, keeping it small.

Migrations are tracked in a `schema_migrations` table (`filename TEXT PRIMARY KEY`).
`RunMigrations` reads all `*.sql` files embedded in the binary, skips already-applied ones,
and runs the rest in filename order.

## Middleware

### Logger (`middleware.Logger`)

Runs on every request:

- Generates a UUID request ID stored as `requestID` in the Gin context.
- Creates a child `slog.Logger` with `request_id`, `method`, and `path` fields, stored as `logger`.
- After the handler returns, logs status code and duration at `INFO` level.

Handlers retrieve the logger with `c.Get("logger")` so all log lines share the same request ID.

### CORS (`middleware.CORS`)

Wraps `gin-contrib/cors`. Configured in `main.go` from env vars:

- `ALLOWED_ORIGINS` — comma-separated list of allowed origins (e.g. `http://localhost:5173`).
- Allowed methods: `GET`, `POST`, `DELETE`, `OPTIONS`.
- Allowed headers: `Origin`, `Content-Length`, `Content-Type`.
- Credentials are disabled.

If `ALLOWED_ORIGINS` is empty, no origins are allowed — set it explicitly in `.env`.

## Key dependencies

| Package | Purpose |
|---------|---------|
| `github.com/gin-gonic/gin` | HTTP router and middleware chain |
| `github.com/gin-contrib/cors` | CORS middleware for Gin |
| `github.com/jackc/pgx/v5` | PostgreSQL driver + connection pool |
| `github.com/matoous/go-nanoid/v2` | URL-safe random ID generation |
| `github.com/google/uuid` | Request ID generation in the logger middleware |
| `github.com/joho/godotenv` | Loads `.env` in development |

## Environment variables

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | PostgreSQL connection string |
| `PORT` | No | Listen port, without colon — e.g. `5555` (prefixed automatically) |
| `ENV` | No | Set to `production` for JSON-formatted logs |
| `LOG_LEVEL` | No | `DEBUG`, `INFO`, `WARN`, or `ERROR` (default `INFO`) |
| `ALLOWED_ORIGINS` | Yes | Comma-separated CORS origins, e.g. `http://localhost:5173` |
| `TRUSTED_PROXIES` | No | Comma-separated list of trusted reverse proxy IPs |
