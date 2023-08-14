package model

import (
	"time"

	"github.com/google/uuid"
)

// Model is a base type that contains shared fields for all model types.
// Usually you don't need to interact with this type directly.
type Model struct {
	// The unique identifier for this instance.
	UUID uuid.UUID

	// These dates will only be set to non-zero values if the instance has been created or soft-deleted respectively.
	// These fields can be set manually, but should be considered read-only in most cases.
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

// Deleted indicates whether this instance is currently soft-deleted.
func (m *Model) Deleted() bool {
	return !m.DeletedAt.IsZero()
}
