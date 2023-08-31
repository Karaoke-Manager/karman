package service

import (
	"errors"
)

// ErrNotFound indicates that the requested entity was not found.
var ErrNotFound = errors.New("not found")
