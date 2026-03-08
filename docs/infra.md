# Infrastructure

All containerisation lives under `infra/docker/`. Two Compose files work together through
Docker Compose's override mechanism: `docker-compose.yml` is the production-shaped base and
`docker-compose.dev.yml` overrides it for local development.

## Directory layout

```
infra/
└── docker/
    ├── docker-compose.yml      Production-shaped base — builds images, exposes ports
    ├── docker-compose.dev.yml  Dev override — live-reload, volume mounts, no image builds
    └── .env                    Shared env vars consumed by both Compose files
```

The `backend/` and `frontend/` Dockerfiles live alongside their source:

```
backend/
├── Dockerfile        Multi-stage production build (builder → alpine runtime)
└── Dockerfile.dev    Minimal Go image with Air installed (no build step)

frontend/
└── Dockerfile        Multi-stage production build (node builder → nginx runtime)
```

## Services

### `api` — Go backend

| Mode | Image source |
|------|-------------|
| Production | Built from `backend/Dockerfile` (multi-stage, final image is Alpine) |
| Development | Built from `backend/Dockerfile.dev` — installs Air, mounts source as a volume, runs `air -c .air.toml` for live reload |

Port `8080` is exposed in both modes. The service declares a healthcheck against `GET /ping`
and other services use `condition: service_healthy` to sequence startup correctly.

### `frontend` — React SPA

| Mode | Image source |
|------|-------------|
| Production | Built from `frontend/Dockerfile` — pnpm build → nginx serving `dist/` |
| Development | `node:20-alpine` image pulled directly; source mounted as a volume; `pnpm dev` runs inside the container with `--host 0.0.0.0` |

Production port: `3000 → 80`. Development port: `5173 → 5173`.

The nginx config (`frontend/nginx.conf`) serves the SPA and falls back to `index.html` for
unknown paths so client-side routing works correctly.

### `db` — PostgreSQL 15

Uses the official `postgres:15` image. Data is persisted in the `db_data` named volume.

| Setting | Value |
|---------|-------|
| User | `shortener_user` |
| Password | `password` (override in `.env` for any shared environment) |
| Database | `shortener` |
| Host port | `5433` (mapped to avoid conflicts with a locally running Postgres on `5432`) |

Healthcheck: `pg_isready -U shortener_user -d shortener`.

### `cache` — Redis

Uses the official `redis` image with no persistence configured (cache-only use case).

| Setting | Value |
|---------|-------|
| Host port | `6379` |

Healthcheck: `redis-cli ping`.

## Port map

| Service | Host port | Container port |
|---------|-----------|----------------|
| Backend API | 8080 | 8080 |
| Frontend (prod) | 3000 | 80 |
| Frontend (dev) | 5173 | 5173 |
| PostgreSQL | 5433 | 5432 |
| Redis | 6379 | 6379 |

## Startup order

```
db (healthy) ──┐
               ├──▶ api (healthy) ──▶ frontend
cache (healthy)┘
```

Dev mode relaxes the frontend dependency to `service_started` so the Vite server can come up
while the API is still initialising.

## Volumes

| Volume | Purpose |
|--------|---------|
| `db_data` | PostgreSQL data directory — persists across `down`/`up` cycles |
| `go-mod-cache` | Go module cache shared into the dev API container |
| `node-modules` | `node_modules` for the dev frontend container |

`go-mod-cache` and `node-modules` only exist in the dev Compose file.

## Environment variables (`.env`)

The `.env` file at `infra/docker/.env` is loaded by Docker Compose and injected into the
`api` service. Edit it before starting the stack.

| Variable | Example value | Description |
|----------|---------------|-------------|
| `PORT` | `8080` | Port the Go server listens on inside the container |
| `BASE_URL` | `http://localhost:3000` | Public base URL used to build `short_url` in API responses |
| `DATABASE_URL` | `postgres://shortener_user:password@db:5432/shortener` | PostgreSQL connection string (`db` = Docker service name) |
| `ENV` | `development` | `production` switches Go logs to JSON format |
| `LOG_LEVEL` | `debug` | `DEBUG` \| `INFO` \| `WARN` \| `ERROR` |
| `ALLOWED_ORIGINS` | `http://localhost:3000,http://localhost:8080` | Comma-separated CORS origins accepted by the backend |
| `TRUSTED_PROXIES` | `http://localhost:3000` | Trusted reverse proxy IPs passed to Gin |
| `REDIS_URL` | `redis://cache:6379` | Redis connection string (`cache` = Docker service name) |
| `REDIS_MAX_RETRIES` | `5` | Maximum Redis command retries on transient errors |
| `REDIS_MIN_RETRY_BACKOFF` | `100ms` | Minimum backoff between Redis retries |
| `REDIS_MAX_RETRY_BACKOFF` | `1s` | Maximum backoff between Redis retries |

> `BASE_URL` must match the public address of the **backend** — it is embedded in the
> `short_url` field of API responses so clients can construct working redirect links.
> In development this is typically the frontend origin because the Vite proxy forwards
> `/:code` requests to the backend transparently.

## Makefile targets

All Compose commands are wrapped in a `Makefile` at the repo root. Run `make help` for a
summary. Common targets:

| Target | Command | Description |
|--------|---------|-------------|
| `make up` | `docker compose -f … up -d` | Start the production stack in the background |
| `make down` | `docker compose -f … down` | Stop and remove containers |
| `make dev` | `docker compose -f … -f … up -d` | Start with dev overrides (live reload) |
| `make dev-build` | build + up with dev overrides | Rebuild images then start dev stack |
| `make build` | `docker compose … build` | Rebuild production images |
| `make restart` | `docker compose … restart` | Restart all running containers |
| `make logs` | `docker compose … logs -f` | Tail all service logs |
| `make clean` | `down -v --remove-orphans` + rmi | Remove containers, volumes, and images |
| `make up-deps` | up `db` + `redis` only | Start dependency services without the app |
| `make down-deps` | stop `db` + `redis` | Stop only the dependency services |
| `make test-unit` | `go tool cover …` | Generate HTML coverage report from existing unit run |
| `make test-integration` | up-deps + go test + down-deps | Run integration tests against live services |

## Dockerfiles

### `backend/Dockerfile` (production)

```
golang:1.25-alpine  →  go mod download  →  go build -o server ./cmd/server
                                                        ↓
                                          alpine:latest  COPY server  CMD ./server
```

CGO is disabled so the binary is fully static and runs on a minimal Alpine image with no
Go toolchain present.

### `backend/Dockerfile.dev` (development)

```
golang:1.25-alpine  →  go install air@latest
```

No application code is copied — the source directory is bind-mounted at runtime and Air
watches it for changes, recompiling and restarting the server automatically.

### `frontend/Dockerfile` (production)

```
node:20-alpine  →  pnpm install --frozen-lockfile  →  pnpm build
                                                           ↓
                                        nginx:alpine  COPY dist/  COPY nginx.conf
```

The nginx config handles SPA routing (unknown paths → `index.html`).
