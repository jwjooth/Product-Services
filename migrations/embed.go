package migrations

import "embed"

// Files contains the SQL migrations used by golang-migrate.
//
//go:embed *.sql
var Files embed.FS
