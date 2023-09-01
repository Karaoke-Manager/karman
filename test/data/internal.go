//go:build database

package testdata

import (
	"time"

	"github.com/google/uuid"
)

// creationResult is the typical result of an INSERT INTO query.
type creationResult struct {
	ID        int
	UUID      uuid.UUID
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
