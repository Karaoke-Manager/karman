// Package migrations contains the database migrations of the Karman project.
// These migrations must be applied before Karman can function correctly.
//
// This Go package is very minimal as migrations are typically written directly in SQL.
// However, since goose does support Go migrations, it could be possible that more complicated migrations will
// be written in Go in the future.
package migrations

import (
	"embed"
)

// FS provides access to all SQL migration files.
// The typical use for this field is in goose.SetBaseFS(migrations.FS)
// which allows Goose to find the migrations in the root "." directory.
//
//go:embed *.sql
var FS embed.FS
