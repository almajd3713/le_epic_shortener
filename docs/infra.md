# Infrastructure

All infrastructure lives under `infra/`. There are three layers: Docker Compose for local dev,
Kubernetes manifests for direct cluster deployment, and Terraform for declarative cluster management.

## Directory layout

```
infra/
├── docker/
│   ├── docker-compose.yml      Production-shaped base — builds images, exposes ports
│   ├── docker-compose.dev.yml  Dev override — live-reload, volume mounts, no image builds
│   └── .env                    Shared env vars consumed by both Compose files
├── k8s/
│   ├── namespace.yaml          shortener namespace
│   ├── configmap.yaml          App env vars (PORT, BASE_URL, REDIS_*, etc.)
│   ├── secrets.yaml            Database URL + Postgres credentials
│   ├── ingress.yaml            NGINX Ingress — routes shortener.local to frontend/api
│   ├── api/
│   │   ├── deployment.yaml     Go API deployment
│   │   └── service.yaml        ClusterIP service on port 8080
│   ├── nginx/
│   │   ├── configmap.yaml      nginx.conf injected into the frontend container
│   │   ├── deployment.yaml     React SPA + nginx deployment
│   │   └── service.yaml        ClusterIP service on port 80
│   ├── postgres/
│   │   ├── statefulset.yaml    PostgreSQL StatefulSet with volumeClaimTemplate
│   │   ├── service.yaml        Headless service (clusterIP: None) named "db"
│   │   └── pvc.yaml            Standalone PVC (used by raw-manifest path only)
│   └── redis/
│       ├── deployment.yaml     Redis deployment
│       └── service.yaml        ClusterIP service named "cache"
└── terraform/
    ├── main.tf                 Root — namespace, configmap, secret, ingress, module calls
    ├── variables.tf            Input variables with defaults
    ├── outputs.tf
    └── modules/
        ├── api/                Go API deployment + service
        ├── nginx/              Frontend deployment + service + nginx ConfigMap
        ├── postgres/           PostgreSQL StatefulSet + headless service
        └── redis/              Redis deployment + service
```

The `backend/` and `frontend/` Dockerfiles live alongside their source:

```
backend/
├── Dockerfile        Multi-stage production build (builder → alpine runtime)
└── Dockerfile.dev    Minimal Go image with Air installed (no build step)

frontend/
└── Dockerfile        Multi-stage production build (node builder → nginx runtime)
```

---

## Docker Compose

### Services

#### `api` — Go backend

| Mode | Image source |
|------|-------------|
| Production | Built from `backend/Dockerfile` (multi-stage, final image is Alpine) |
| Development | Built from `backend/Dockerfile.dev` — installs Air, mounts source as a volume, runs `air -c .air.toml` for live reload |

Port `8080` is exposed in both modes. The service declares a healthcheck against `GET /ping`
and other services use `condition: service_healthy` to sequence startup correctly.

#### `frontend` — React SPA

| Mode | Image source |
|------|-------------|
| Production | Built from `frontend/Dockerfile` — pnpm build → nginx serving `dist/` |
| Development | `node:20-alpine` image pulled directly; source mounted as a volume; `pnpm dev` runs inside the container with `--host 0.0.0.0` |

Production port: `3000 → 80`. Development port: `5173 → 5173`.

The nginx config (`frontend/nginx.conf`) serves the SPA and proxies `/api/` and unknown paths
to the Go backend.

#### `db` — PostgreSQL 15

Uses the official `postgres:15` image. Data is persisted in the `db_data` named volume.

| Setting | Value |
|---------|-------|
| User | `shortener_user` |
| Password | `password` (override in `.env` for any shared environment) |
| Database | `shortener` |
| Host port | `5433` (mapped to avoid conflicts with a locally running Postgres on `5432`) |

Healthcheck: `pg_isready -U shortener_user -d shortener`.

#### `cache` — Redis

Uses the official `redis` image with no persistence configured (cache-only use case).

| Setting | Value |
|---------|-------|
| Host port | `6379` |

Healthcheck: `redis-cli ping`.

#### `prometheus` — Metrics scraper

Uses `prom/prometheus` with a local config mounted from
`infra/docker/prometheus/prometheus.yml`.

| Setting | Value |
|---------|-------|
| Host port | `9090` |
| Scrape target | `api:8080/metrics` |

Prometheus stores time-series data in the `prometheus_data` named volume.

#### `grafana` — Metrics dashboard

Uses `grafana/grafana` with provisioning mounted from
`infra/docker/grafana/provisioning/`.

| Setting | Value |
|---------|-------|
| Host port | `3001` (container `3000`) |
| Default user | `admin` (`GRAFANA_ADMIN_USER`) |
| Default password | `admin` (`GRAFANA_ADMIN_PASSWORD`) |

A Prometheus datasource is provisioned automatically and points to
`http://prometheus:9090`. Grafana data is persisted in `grafana_data`.

### Port map

| Service | Host port | Container port |
|---------|-----------|----------------|
| Backend API | 8080 | 8080 |
| Frontend (prod) | 3000 | 80 |
| Frontend (dev) | 5173 | 5173 |
| PostgreSQL | 5433 | 5432 |
| Redis | 6379 | 6379 |
| Prometheus | 9090 | 9090 |
| Grafana | 3001 | 3000 |

### Startup order

```
db (healthy) ──┐
               ├──▶ api (healthy) ──▶ frontend
cache (healthy)┘

api (healthy) ──▶ prometheus ──▶ grafana
```

Dev mode relaxes the frontend dependency to `service_started` so the Vite server can come up
while the API is still initialising.

### Volumes

| Volume | Purpose |
|--------|---------|
| `db_data` | PostgreSQL data directory — persists across `down`/`up` cycles |
| `prometheus_data` | Prometheus TSDB storage |
| `grafana_data` | Grafana dashboards, users, and settings |

---

## Kubernetes

### Prerequisites

- Minikube or k3s with the NGINX Ingress Controller enabled
- `shortener.local` pointing to your cluster IP in `/etc/hosts`

```bash
minikube addons enable ingress
echo "$(minikube ip)  shortener.local" | sudo tee -a /etc/hosts
```

### Workloads

| Workload | Kind | Name | Notes |
|----------|------|------|-------|
| Go API | Deployment | `api` | `envFrom` loads ConfigMap + Secret; readiness on `GET /ping` |
| React frontend | Deployment | `frontend` | nginx.conf injected via ConfigMap |
| PostgreSQL | StatefulSet | `shortener` | `volumeClaimTemplate` with `local-path` StorageClass |
| Redis | Deployment | `cache` | No persistence; readiness via `redis-cli ping` |

### Services

| Service | Type | Name | Port | Target |
|---------|------|------|------|--------|
| Go API | ClusterIP | `api` | 8080 | API pods |
| Frontend | ClusterIP | `frontend` | 80 | Frontend pods |
| PostgreSQL | Headless (ClusterIP: None) | `db` | 5432 | StatefulSet pods |
| Redis | ClusterIP | `cache` | 6379 | Redis pods |

Service names are meaningful — `REDIS_URL: redis://cache:6379` and `DATABASE_URL: ...@db:5432/...`
resolve directly via Kubernetes DNS.

### Ingress

A single `kubernetes.io/ingress` resource (`shortener-ingress`) routes `shortener.local`:

| Path | Backend service | Port |
|------|----------------|------|
| `/api` (Prefix) | `api` | 8080 |
| `/` (Prefix) | `frontend` | 80 |

Short-code redirects (`GET /:code`) hit the frontend nginx first, which falls back to the API
via the `@backend` location block.

### ConfigMap keys

All keys in `shortener-config` map directly to the environment variable names that `config.go` reads:

| Key | Default | Description |
|-----|---------|-------------|
| `PORT` | `8080` | API listen port |
| `BASE_URL` | `http://shortener.local` | Prefix for generated short URLs |
| `ENV` | `production` | Logging mode (`development` = text, `production` = JSON) |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `ALLOWED_ORIGINS` | `http://shortener.local` | CORS allowed origins (comma-separated) |
| `TRUSTED_PROXIES` | `` | Gin trusted proxy IPs (comma-separated) |
| `REDIS_URL` | `redis://cache:6379` | Redis connection string |
| `REDIS_MAX_RETRIES` | `5` | Max Redis reconnect attempts |
| `REDIS_MIN_RETRY_BACKOFF` | `100ms` | Min delay between retries |
| `REDIS_MAX_RETRY_BACKOFF` | `1s` | Max delay between retries |

### Secret keys

`shortener-secrets` holds:

| Key | Description |
|-----|-------------|
| `database_url` | Full Postgres connection string for the Go API |
| `postgres_user` | Postgres username — also used by the StatefulSet container |
| `postgres_password` | Postgres password |
| `postgres_db` | Postgres database name |

---

## Terraform

Terraform manages the same cluster resources as the raw manifests, using the `hashicorp/kubernetes`
provider with `WaitForDefaultServiceAccount = false` (no cloud account required).

### Modules

| Module | Path | Manages |
|--------|------|---------|
| `postgres` | `modules/postgres` | StatefulSet, headless service; PVC via `volume_claim_template` |
| `redis` | `modules/redis` | Deployment, ClusterIP service |
| `api` | `modules/api` | Deployment (with `envFrom`), ClusterIP service |
| `nginx` | `modules/nginx` | Deployment, ClusterIP service, nginx ConfigMap |

Root `main.tf` manages: namespace, shared ConfigMap, shared Secret, and Ingress resource.

### Key variables (`variables.tf`)

| Variable | Default | Description |
|----------|---------|-------------|
| `base_url` | `http://shortener.local` | Injected into ConfigMap as `BASE_URL` |
| `allowed_origins` | `http://shortener.local` | Injected into ConfigMap as `ALLOWED_ORIGINS` |
| `app_port` | `8080` | API container port |
| `environment` | `production` | Injected as `ENV` |
| `log_level` | `info` | Injected as `LOG_LEVEL` |
| `database_url` | — | Full Postgres connection string |
| `postgres_user` | — | Postgres username |
| `postgres_password` | — | Postgres password (sensitive) |
| `postgres_db` | — | Postgres database name |
| `storage_class` | `local-path` | StorageClass for the Postgres PVC |

### Usage

```bash
cd infra/terraform
terraform init
terraform apply          # uses variable defaults
```

Override values in `terraform.tfvars`:

```hcl
base_url          = "http://shortener.local"
allowed_origins   = "http://shortener.local"
postgres_password = "changeme"
```

### StorageClass note

The default `local-path` StorageClass uses `VolumeBindingMode: WaitForFirstConsumer`. The Postgres
StatefulSet uses a `volume_claim_template` (not a standalone PVC resource) so that the PVC is
created at pod-scheduling time — avoiding a deadlock where Terraform waits on a PVC that can't
bind until a pod exists.

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
| `GRAFANA_ADMIN_USER` | `admin` | Grafana admin username |
| `GRAFANA_ADMIN_PASSWORD` | `admin` | Grafana admin password |

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
| `make up-deps` | up `db` + `redis` + `prometheus` + `grafana` | Start dependency and observability services without the app |
| `make down-deps` | stop `db` + `redis` + `prometheus` + `grafana` | Stop dependency and observability services |
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

---

## CI/CD

Pipelines live under `.github/workflows/`.

### CI Pipeline (`.github/workflows/ci.yml`)

Runs on every **pull request targeting `main`**. Three jobs run in parallel:

| Job | What it does |
|-----|-------------|
| `backend-unit-tests` | Checks out, sets up Go 1.25 (with module-cache caching), runs `make test-unit` |
| `backend-integration-tests` | Spins up Postgres 15 and Redis 7 as service containers, then runs `go test -v -race -tags=integration ./...` directly inside `backend/` |
| `frontend` | Sets up Node 25 and pnpm 10 (with pnpm-store caching), installs deps with `--frozen-lockfile`, then runs `pnpm build` |

The integration-test job uses the same credentials as the local dev stack:

| Service | Config |
|---------|--------|
| Postgres | `shortener_user / password / shortener`, health-checked with `pg_isready` |
| Redis | Default port `6379`, health-checked with `redis-cli ping` |

### CD Pipeline (`.github/workflows/cd.yml`)

Triggered on any **tag push matching `phase-*`** (e.g. `phase-05`). Requires `contents: write` and `packages: write` permissions so it can push images to GHCR and create GitHub Releases.

A `concurrency` group (`cd-pipeline`, `cancel-in-progress: false`) prevents two releases from racing; a second tag push queues rather than cancels.

**Job graph:**

```
build-backend ──┐
                ├──▶ verify ──▶ create-release
build-frontend ─┘
```

#### `build-backend` / `build-frontend`

Both jobs follow the same pattern:

1. Set up Docker Buildx.
2. Log in to **GHCR** (`ghcr.io`) using `GITHUB_TOKEN`.
3. Use `docker/metadata-action` to produce two tags: the short commit SHA and `latest` (on the default branch).
4. Build and push with `docker/build-push-action`, passing `GIT_COMMIT`, `GIT_REF`, `GIT_TAG`, and `BUILD_DATE` as build-args so the binary can embed version information.
5. Layer cache is stored in GitHub Actions cache (`type=gha`).

| Image | Registry path |
|-------|---------------|
| Backend | `ghcr.io/<owner>/<repo>/api` |
| Frontend | `ghcr.io/<owner>/<repo>/frontend` |

#### `verify`

Pulls and runs `docker inspect` on both `latest` images to confirm they were pushed successfully before a release is created.

#### `create-release`

1. Fetches the full tag history (`fetch-depth: 0`).
2. Finds the previous tag with `git describe --tags --abbrev=0`.
3. Generates a changelog from `git log <prev-tag>..HEAD --pretty=format:"- %s (%h)"`.
4. Creates a GitHub Release via `softprops/action-gh-release@v2` with the tag name, a human-readable title, and the generated changelog as the body.
