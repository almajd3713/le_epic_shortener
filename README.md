# URL Shortener

A self-hosted URL shortener built with Go and PostgreSQL.

## Stack

- **Backend** — Go, Gin, pgx
- **Database** — PostgreSQL
- **Dev tooling** — Air (live reload), Docker Compose

## Running locally

**Prerequisites:** Go 1.25+, a running PostgreSQL instance.

1. Copy `.env.example` to `.env` inside `backend/` and fill in the values:

   ```env
   DATABASE_URL=postgres://user:password@localhost:5432/shortener
   PORT=5555
   ENV=development       # set to "production" for JSON logs
   LOG_LEVEL=INFO        # DEBUG | INFO | WARN | ERROR
   ```

2. Start the server:

   ```bash
   cd backend
   go run ./cmd/server
   ```

   Or with live reload:

   ```bash
   air
   ```

Migrations run automatically on startup.

## API

| Method | Path              | Description                              |
|--------|-------------------|------------------------------------------|
| `GET`  | `/ping`           | Health check — returns `{"message":"pong"}` |
| `POST` | `/api/shorten`    | Shorten a URL                            |
| `GET`  | `/:code`          | Redirect to the original URL             |

### POST `/api/shorten`

**Request body**

```json
{
  "long_url": "https://example.com/very/long/path",
  "expires_at": "2026-12-31T00:00:00Z"   // optional
}
```

**Response**

```json
{
  "shortened_url": "axsnZ_BX"
}
```

### GET `/:code`

Responds with `302 Found` and a `Location` header pointing to the original URL.
Returns `404` if the code is unknown or expired.
