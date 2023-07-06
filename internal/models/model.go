package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Model struct {
	gorm.Model
	UUID uuid.UUID `gorm:"type:uuid,uniqueIndex"`
}

func (m *Model) BeforeCreate(tx *gorm.DB) error {
	m.UUID = uuid.New()
	return nil
}
