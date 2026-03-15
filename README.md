# LinkPulse

A professional, portfolio-style URL shortener service written in Go.

LinkPulse demonstrates clean layered architecture, practical backend design, async click tracking, a server-rendered dashboard UI, and real-time analytics streaming.

---

## Table of Contents

- [Overview](#overview)
- [Feature Set](#feature-set)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Services in This Repository](#services-in-this-repository)
- [Runtime Configuration](#runtime-configuration)
- [How It Works](#how-it-works)
- [HTTP API](#http-api)
- [Dashboard and UI](#dashboard-and-ui)
- [Metrics and Observability](#metrics-and-observability)
- [Run Locally](#run-locally)
- [Run with Docker Compose](#run-with-docker-compose)
- [Testing](#testing)
- [Security Notes](#security-notes)
- [Known Limitations](#known-limitations)
- [Roadmap / Production Hardening](#roadmap--production-hardening)

---

## Overview

LinkPulse is a URL shortener that supports:

- Short URL creation with generated codes or custom aliases.
- Optional expiration for links.
- Optional password-protected access.
- Redirect handling with click tracking.
- Real-time click feed via server-sent events (SSE).
- A responsive dashboard for link management and analytics.

The current implementation is intentionally self-contained and uses in-memory storage adapters for easy local execution.

---

## Feature Set

### Core Link Features
- Create short links from long URLs.
- Custom alias validation (`[a-zA-Z0-9_-]{4,20}`).
- Optional expiration timestamp.
- Optional password protection.

### Redirect & Tracking
- Resolve short code to destination URL.
- Expiration guard (returns HTTP 410 when expired).
- Password gate page for protected links.
- Asynchronous click tracking (non-blocking redirect path).
- Click metadata captured: referrer, user-agent, IP, parsed browser/OS.

### Analytics
- Total and unique click counters.
- Top referrers and top browsers.
- Live feed updates on link detail pages via SSE stream.

### UI
- Dashboard page for creating and listing links.
- Link detail analytics page with live updates.
- Password unlock page with polished UX.
- HTMX + Alpine.js powered interactions.

### Operations
- Lightweight `/metrics` endpoint.
- Graceful server shutdown on SIGINT/SIGTERM.
- Dockerfile + docker-compose scaffolding.
- Prometheus and Grafana compose services included.

---

## Architecture

LinkPulse follows a layered design:

1. **Presentation layer**
   - HTTP handlers and templates.
   - Route handling, request/response serialization, HTML rendering.

2. **Service layer**
   - Business rules (validation, password checks, tracking orchestration).
   - Code generation and domain-level decision logic.

3. **Repository layer**
   - Data access contracts + in-memory implementation.
   - Analytics aggregation from stored click events.

4. **Cache layer**
   - In-memory cache for link lookup acceleration.

5. **Event hub layer**
   - Observer-style pub/sub hub for streaming click events to subscribers.

### Patterns Used
- Repository pattern.
- Service layer pattern.
- Observer-like event broadcasting.
- Middleware pattern for cross-cutting concerns.

---

## Project Structure

```text
cmd/server/main.go                    # App entrypoint and wiring
internal/config/                      # Runtime config loader
internal/logger/                      # Logger setup
internal/models/                      # Domain models
internal/repository/                  # Interfaces + in-memory repo
internal/cache/                       # In-memory cache adapter
internal/service/                     # Business logic
internal/http/handlers/               # HTTP handlers + routing
internal/http/middleware/             # Middleware (metrics wrapper)
internal/ws/                          # SSE hub/pub-sub
internal/metrics/                     # In-process counters
web/templates/                        # Dashboard, detail, password templates
deployments/prometheus.yml            # Prometheus scrape config
deployments/grafana/*.json            # Example Grafana dashboard JSON
Dockerfile                            # App container image
docker-compose.yml                    # App + observability + db/cache scaffold
```

---

## Services in This Repository

`docker-compose.yml` defines the following services:

1. **app**
   - The Go LinkPulse service (port `8080`).
   - Current runtime uses in-memory repository/cache.

2. **postgres**
   - PostgreSQL container (port `5432`).
   - Present as infrastructure scaffolding for future persistent adapter integration.

3. **redis**
   - Redis container (port `6379`).
   - Present as infrastructure scaffolding for future cache/counter integration.

4. **prometheus**
   - Scrapes metrics from the app (`app:8080`).
   - Exposed on port `9090`.

5. **grafana**
   - Visualization UI exposed on port `3000`.
   - Includes sample dashboard JSON in `deployments/grafana`.

> Note: In the current code, Postgres/Redis adapters are placeholders; primary behavior is backed by memory implementations.

---

## Runtime Configuration

The app reads configuration from environment variables with sane defaults:

| Variable | Default | Description |
|---|---|---|
| `HTTP_PORT` | `8080` | HTTP listen port |
| `BASE_URL` | `http://localhost:8080` | Public base URL used in generated short links |

You can set them before running:

```bash
export HTTP_PORT=8080
export BASE_URL=http://localhost:8080
```

---

## How It Works

### 1) Create Link
- Client submits URL + optional alias/expiry/password.
- Service validates input and builds a short code.
- Password (if provided) is hashed before storage.
- Link is written to repository and cached.

### 2) Resolve + Redirect
- Incoming `GET /{code}` resolves from cache/repository.
- Expired links are rejected.
- Protected links require password unlock.
- Click tracking is dispatched asynchronously.
- Client is redirected to destination URL.

### 3) Click Tracking + Analytics
- Click metadata is recorded in repository.
- Unique-visitor approximation uses IP + user-agent keying.
- Aggregations provide total/unique and top dimensions.
- Live event is published to SSE subscribers for the link.

---

## HTTP API

### REST-style endpoints

#### Create a short link
`POST /api/links`

Request body:

```json
{
  "long_url": "https://golang.org",
  "custom_alias": "golang",
  "expires_at": "2027-01-01T00:00:00Z",
  "password": "secret"
}
```

Response includes created link and `short_url`.

#### List links
`GET /api/links`

#### Get link by code
`GET /api/links/{code}`

#### Get analytics summary
`GET /api/links/{code}/clicks`

#### Delete link
`DELETE /api/links/{code}`

### Redirect and access
- `GET /{code}` — redirect flow with expiry/password checks.
- `POST /access/{code}` — submit password for protected links.

### Real-time stream
- `GET /stream/{code}` — server-sent events stream for click updates.

---

## Dashboard and UI

### Pages
- `/` — dashboard (create form + links table).
- `/links/{code}` — analytics details + live click feed.
- Password page rendered when protected link access is attempted.

### Frontend stack
- Server-rendered HTML templates.
- HTMX for form update UX.
- Alpine.js for small client-side interactions.
- Custom CSS for responsive, polished visuals.

---

## Metrics and Observability

### App endpoint
- `GET /metrics`

Current output is a simple text payload with:
- links created
- total clicks
- active stream connections

### Prometheus
Configured via `deployments/prometheus.yml` to scrape `app:8080`.

### Grafana
Sample dashboard JSON is included in:
- `deployments/grafana/linkpulse-dashboard.json`

---

## Run Locally

### Prerequisites
- Go 1.23+

### Start app

```bash
go run ./cmd/server
```

Open:
- App: http://localhost:8080
- Metrics: http://localhost:8080/metrics

---

## Run with Docker Compose

```bash
docker compose up --build
```

Exposed services:
- App: http://localhost:8080
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000
- Postgres: localhost:5432
- Redis: localhost:6379

---

## Testing

Run all tests:

```bash
go test ./...
```

Build validation:

```bash
go build ./...
```

Service tests cover:
- Link creation and resolution.
- Password validation.
- Async click tracking path.

---

## Security Notes

- Password-protected links are supported, and passwords are hashed before storage.
- Alias and URL inputs are validated in service layer.
- Cookies used for access are marked `HttpOnly`.

For production, consider:
- stronger password hashing policy (bcrypt/argon2),
- HTTPS-only and secure cookie flags,
- stricter input normalization,
- authN/authZ for dashboard/API,
- rate-limiting and abuse controls.

---

## Known Limitations

- In-memory storage only (data resets on restart).
- SSE is used for live updates (WebSocket endpoint not currently implemented).
- Basic metrics format instead of full Prometheus client instrumentation.
- No user accounts or multi-tenant ownership model.
- Docker Compose includes infra services that are scaffolding for future adapters.

---

## Roadmap / Production Hardening

1. Add real PostgreSQL repository adapter and migrations.
2. Add Redis cache/counter adapter and invalidation strategy.
3. Replace simple password hashing with bcrypt/argon2.
4. Add authentication and per-user link ownership.
5. Add pagination/filtering/sorting across APIs and UI.
6. Add richer analytics (time windows, geo, UA parsing library).
7. Expand observability with structured logs and Prometheus metrics library.
8. Add integration tests + end-to-end tests in CI.

---

If you are evaluating this project as a portfolio artifact, the codebase is intentionally organized to make these upgrades straightforward while already demonstrating practical architecture and implementation discipline.
