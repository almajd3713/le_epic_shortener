# URL Shortener

A self-hosted URL shortener built with Go, PostgreSQL, and React.

## Stack

- **Backend** — Go, Gin, pgx
- **Frontend** — React 19, Vite, Tailwind CSS v4, TypeScript
- **Database** — PostgreSQL
- **Cache** — Redis
- **Dev tooling** — Air (live reload), pnpm
- **Containerization** — Docker, Docker Compose

## Running locally

**Prerequisites:** Go 1.25+, Node.js 20+, pnpm, a running PostgreSQL instance.

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

The dev server starts at `http://localhost:5173` and proxies `/api/*` and `/:code` redirects to the Go backend. The target defaults to `http://localhost:8080`; override it by setting `VITE_API_URL` in `frontend/.env.local`.

## Running with Docker

**Prerequisites:** Docker with the Compose plugin.

```bash
make dev        # development stack (live reload, volume mounts)
make up         # production-shaped stack
```

See [docs/infra.md](docs/infra.md) for the full service map, port assignments, environment
variables, Dockerfile descriptions, and all available `make` targets.

## API

| Method | Path                  | Description                                      |
|--------|-----------------------|--------------------------------------------------|
| `GET`  | `/ping`               | Health check — returns `{"message":"pong"}`      |
| `POST` | `/api/shorten`        | Shorten a URL                                    |
| `GET`  | `/api/urls`           | List all shortened URLs (all states)             |
| `GET`  | `/geturl/:code`       | Return the original long URL for a short code    |
| `GET`  | `/:code`              | Redirect to the original URL                     |
| `PATCH`| `/api/toggle/:code`   | Activate or deactivate a shortened URL           |
| `DELETE`| `/api/delete/:code`  | Permanently delete a shortened URL               |

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
  "short_url": "http://localhost:8080/axsnZ_BX",
  "created_at": "2026-03-06T12:00:00Z"
}
```

### GET `/api/urls`

Returns a JSON array of **all** shortened URLs (active and inactive, including expired) ordered by
creation date (newest first).

### GET `/geturl/:code`

Returns the original long URL for a short code without triggering a redirect.

**Response**

```json
{ "long_url": "https://example.com/very/long/path" }
```

Returns `404` if the code is unknown.

### PATCH `/api/toggle/:code`

Activates or deactivates a shortened URL.

**Request body**

```json
{ "action": "activate" }   // or "deactivate"
```

### DELETE `/api/delete/:code`

Permanently removes a shortened URL from the database.

### GET `/:code`

Responds with `302 Found` and a `Location` header pointing to the original URL.
Returns `404` if the code is unknown, inactive, or expired.
