package dbutil

import (
	"testing"
	"time"
)

func TestZeroNil(t *testing.T) {
	t.Parallel()

	now := time.Now()
	zero := time.Time{}
	actual := ZeroNil(&now)
	if now != actual {
		t.Errorf("ZeroNil(&now) = %s, expected %s", actual, now)
	}

	actual = ZeroNil((*time.Time)(nil))
	if actual != zero {
		t.Errorf("ZeroNil(nil) = %s, expected zero value", actual)
	}
}
