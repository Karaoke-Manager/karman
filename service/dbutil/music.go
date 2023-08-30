package dbutil

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"io"
	"strings"

	"codello.dev/ultrastar"
	"codello.dev/ultrastar/txt"
)

type Music ultrastar.Music

func (m *Music) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	b := &strings.Builder{}
	err := txt.FormatDefault.WriteMusic(b, (*ultrastar.Music)(m))
	return b.String(), err
}

func (m *Music) Scan(value any) error {
	if value == nil {
		return nil
	}
	var r io.Reader
	switch v := value.(type) {
	case []byte:
		r = bytes.NewReader(v)
	case string:
		r = strings.NewReader(v)
	default:
		return fmt.Errorf("invalid type for music: %T", value)
	}
	m2, err := txt.DialectDefault.ReadMusic(r)
	if err != nil {
		return err
	}
	*m = Music(*m2)
	return nil
}
