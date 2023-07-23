package model

import (
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Model is the base type for most Karman models.
// Karman identifies its entities via UUID.
// This type implements the corresponding field.
type Model struct {
	gorm.Model
	UUID uuid.UUID `gorm:"type:uuid,uniqueIndex"`
}

// Deleted indicates whether m is soft-deleted.
func (m *Model) Deleted() bool {
	return m.DeletedAt.Valid
}

// BeforeCreate ensures that m.UUID is set to a valid value.
func (m *Model) BeforeCreate(tx *gorm.DB) error {
	if m.UUID == uuid.Nil {
		m.UUID = uuid.New()
	}
	return nil
}

// BeforeUpdate checks that m.UUID does not change.
func (m *Model) BeforeUpdate(tx *gorm.DB) error {
	if tx.Statement.Changed("UUID") {
		return errors.New("UUID not allowed to change")
	}
	return nil
}
