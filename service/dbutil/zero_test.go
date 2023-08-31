package dbutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestZeroNil(t *testing.T) {
	now := time.Now()
	assert.Equal(t, now, ZeroNil(&now))
	assert.Zero(t, ZeroNil((*time.Time)(nil)))
}
