package common

import (
	"errors"

	"gorm.io/gorm"
)

var (
	ErrNotFound = errors.New("not found")
)

func DBError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	} else {
		return err
	}
}
