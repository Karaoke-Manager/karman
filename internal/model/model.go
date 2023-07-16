package model

import (
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Model struct {
	gorm.Model
	UUID uuid.UUID `gorm:"type:uuid,uniqueIndex"`
}

func (m *Model) Deleted() bool {
	return m.DeletedAt.Valid
}

func (m *Model) BeforeCreate(tx *gorm.DB) error {
	if m.UUID.Version() == 0 {
		m.UUID = uuid.New()
	}
	return nil
}

func (m *Model) BeforeUpdate(tx *gorm.DB) error {
	if tx.Statement.Changed("UUID") {
		return errors.New("UUID not allowed to change")
	}
	return nil
}
