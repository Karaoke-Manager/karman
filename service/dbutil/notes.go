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

type Notes ultrastar.Notes

func (ns Notes) Value() (driver.Value, error) {
	if ns == nil {
		return nil, nil
	}
	b := &strings.Builder{}
	err := txt.NewWriter(b).WriteNotes(ultrastar.Notes(ns))
	return b.String(), err
}

func (ns *Notes) Scan(value any) error {
	if value == nil {
		*ns = nil
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
	ns2, err := txt.NewReader(r).ReadNotes()
	if err != nil {
		return err
	}
	*ns = Notes(ns2)
	return nil
}
