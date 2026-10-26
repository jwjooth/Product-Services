# Project guidance

Go 1.26 product REST API using Chi, GORM, and MySQL. `main.go` wires routes and graceful shutdown.

## Principles

- Make the smallest complete change. Reuse existing patterns; avoid unnecessary abstractions, dependencies, configuration, and refactors.
- Be strict about correctness, security, data integrity, API compatibility, and error handling. Raise only concrete, actionable issues; state impact and the smallest viable fix.
- Keep boundaries clear: `controller/` handles HTTP, `service/` handles validation and business rules, and `repository/` handles GORM and persistence. Do not leak HTTP concerns into services or DTOs into repositories.
- Preserve the existing API and response envelopes unless the request requires a change. `Product.Category` is a string, not a relation.

## Project details

- Routes: `/healthz`, `/swagger/*`, and `/api/v1/products` (paginated list plus create/read/update/delete).
- Tests use mocks and do not require MySQL. `go run .` requires a reachable database.
- `config.Migrate` auto-migrates the product model; use versioned migrations for production schema changes.
- Swagger files in `docs/` may be stale. After changing annotations or payloads, regenerate them with `swag init -g main.go -o docs` and include the changes.
- Environment variables are the production configuration source. Never commit `.env`; `APP_URL` configures the Swagger host and scheme.

## Validation

Run focused tests first, then the checks relevant to the change:

```bash
go test ./...
go vet ./...
go build ./...
gofmt -l .
```

`gofmt -l .` must print no files. The CI workflow also runs golangci-lint.
