package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Karaoke-Manager/karman/model"
)

// Entity is the base type for most Karman models.
// Karman identifies its entities via UUID.
// This type implements the corresponding field.
type Entity struct {
	gorm.Model
	UUID uuid.UUID `gorm:"type:uuid,uniqueIndex"`
}

// fromModel creates a new Entity with data from m.
func fromModel(m model.Model) Entity {
	e := Entity{
		Model: gorm.Model{
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		},
		UUID: m.UUID,
	}
	if !m.DeletedAt.IsZero() {
		e.DeletedAt = gorm.DeletedAt{
			Time:  m.DeletedAt,
			Valid: true,
		}
	}
	return e
}

// toModel converts e into a model instance.
// In contrast to other toModel methods in this package this does not return a pointer.
// This makes it a little simpler to use in other toModel methods.
func (e *Entity) toModel() model.Model {
	deletedAt := time.Time{}
	if e.DeletedAt.Valid {
		deletedAt = e.DeletedAt.Time
	}
	return model.Model{
		UUID:      e.UUID,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
		DeletedAt: deletedAt,
	}
}

// BeforeCreate ensures that m.UUID is set to a valid value.
func (e *Entity) BeforeCreate(tx *gorm.DB) error {
	if e.UUID == uuid.Nil {
		e.UUID = uuid.New()
	}
	return nil
}

// BeforeUpdate checks that m.UUID does not change.
func (e *Entity) BeforeUpdate(tx *gorm.DB) error {
	if tx.Statement.Changed("UUID") {
		return errors.New("UUID not allowed to change")
	}
	return nil
}
