# Product-Services Agent Instructions

## Commands
- **Run server**: `go run .` (runs on port `6767`, health check at `/healthz`)
- **Run tests**: `go test ./...`
- **Swagger generate**: `swag init`

## Environment & Auth
- **Database**: MySQL (`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`).
- **Admin Auth**: `ADMIN_API_KEY` (min 32 chars). POST/PUT/DELETE require header `X-Admin-Key` or `Authorization: Bearer <ADMIN_API_KEY>`.

## Architecture
- `main.go`: App entrypoint & router setup (`github.com/go-chi/chi/v5`).
- `controller/`: HTTP request handlers.
- `service/`: Business logic layer.
- `repository/`: Data access layer (`gorm.io/gorm`).
- `entities/`: DB models.
- `payload/`: Request validation payloads (`github.com/go-playground/validator/v10`).
- `auth/`: API key middleware.
- `test/`: Unit tests mirroring package structure (uses `go-sqlmock` and `stretchr/testify`).
