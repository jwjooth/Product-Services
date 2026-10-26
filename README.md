# Golang Product REST API

A simple product CRUD API built with Go, Chi, GORM, and MySQL.

## Features

- Create, list, view, update, and delete products
- Paginated product listing
- Request validation and Swagger UI

## Run locally

Requirements: Go 1.26.5+ and MySQL.

Set your database connection using `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, and `DB_NAME`. Defaults are `localhost:3306`, user `root`, and database `products`. Set `ADMIN_API_KEY` to an admin key with at least 32 characters.

POST, PUT, and DELETE product endpoints require either `X-Admin-Key: <ADMIN_API_KEY>` or `Authorization: Bearer <ADMIN_API_KEY>`.

```bash
go run .
```

The server runs on port `6767` by default. Check `http://localhost:6767/healthz` to verify it is running.

## API

Base path: `/api/v1`

| Method | Endpoint | Description |
|---|---|---|
| GET | `/products?page=1&per_page=10` | List products |
| POST | `/products` | Create a product |
| GET | `/products/{id}` | Get a product |
| PUT | `/products/{id}` | Update a product |
| DELETE | `/products/{id}` | Delete a product |

Swagger UI: `http://localhost:6767/swagger/index.html`

## Tests

```bash
go test ./...
```

## License

[MIT](LICENSE)
