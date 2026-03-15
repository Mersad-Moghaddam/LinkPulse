# LinkPulse - Advanced URL Shortener

LinkPulse is a layered Go URL shortener that includes custom aliases, optional expiry/password protection, analytics, and a responsive HTMX dashboard.

## Current implementation
- In-memory repository and cache (easy to run locally, no external services required).
- Layered architecture: handlers -> services -> repositories.
- Redirect tracking with asynchronous click recording.
- Real-time analytics updates over **Server-Sent Events** stream.
- Dashboard templates (HTMX + Alpine.js) for link creation and analytics.
- Minimal metrics endpoint (`/metrics`).

## Quick start
```bash
go run ./cmd/server
```

Open http://localhost:8080.

## API
- `POST /api/links`
- `GET /api/links`
- `GET /api/links/{code}`
- `GET /api/links/{code}/clicks`
- `DELETE /api/links/{code}`
- `GET /stream/{code}` (SSE real-time stream)

## Docker Compose
A `docker-compose.yml` is included as a portfolio scaffold and can be used as a baseline when replacing in-memory storage with PostgreSQL/Redis adapters.

## Design patterns
- Repository pattern for data access contracts.
- Service layer for business rules and validation.
- Observer-like hub for streaming click events.
- Middleware for cross-cutting concerns.

## Testing
```bash
go test ./...
```
