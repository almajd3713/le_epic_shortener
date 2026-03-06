# URL Shortener

A self-hosted URL shortener built with Go, PostgreSQL, and React.

## Stack

- **Backend** — Go, Gin, pgx
- **Frontend** — React 19, Vite, Tailwind CSS v4, TypeScript
- **Database** — PostgreSQL
- **Dev tooling** — Air (live reload), pnpm

## Running locally

**Prerequisites:** Go 1.22+, Node.js 20+, pnpm, a running PostgreSQL instance.

### Backend

1. Copy `.env.example` to `.env` inside `backend/` and fill in the values:

   ```env
   DATABASE_URL=postgres://user:password@localhost:5432/shortener
   PORT=5555
   ENV=development         # set to "production" for JSON logs
   LOG_LEVEL=INFO          # DEBUG | INFO | WARN | ERROR
   ALLOWED_ORIGINS=http://localhost:5173
   TRUSTED_PROXIES=        # leave empty for local dev
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

### Frontend

```bash
cd frontend
pnpm install
pnpm dev
```

The dev server starts at `http://localhost:5173` and proxies `/api/*` to the Go backend at `http://localhost:5555`.

## API

| Method | Path              | Description                              |
|--------|-------------------|------------------------------------------|
| `GET`  | `/ping`           | Health check — returns `{"message":"pong"}` |
| `POST` | `/api/shorten`    | Shorten a URL                            |
| `GET`  | `/api/urls`       | List all active shortened URLs           |
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
  "short_code": "axsnZ_BX",
  "shortened_url": "localhost:5555/axsnZ_BX",
  "created_at": "2026-03-06T12:00:00Z"
}
```

### GET `/api/urls`

Returns a JSON array of all active, non-expired URLs ordered by creation date (newest first).

### GET `/:code`

Responds with `302 Found` and a `Location` header pointing to the original URL.
Returns `404` if the code is unknown or expired.
