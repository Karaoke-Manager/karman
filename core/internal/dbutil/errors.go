package dbutil

import (
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/Karaoke-Manager/karman/core"
)

func Error(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return core.ErrNotFound
	}
	return err
}
