package streamio

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"testing/iotest"
)

func TestBufReader(t *testing.T) {
	t.Run("test reader", func(t *testing.T) {
		r := strings.NewReader("This is a hello world string.")
		s := NewBufferedReadSeeker(r, 50)
		err := iotest.TestReader(s, []byte("This is a hello world string."))
		assert.NoError(t, err)
	})
}
