package mediatype

import (
	"database/sql/driver"
	"fmt"
)

// MarshalText serializes t as a string.
func (t MediaType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText deserializes string data.
func (t *MediaType) UnmarshalText(data []byte) (err error) {
	*t, err = Parse(string(data))
	return
}

// MarshalText serializes l as a string.
func (l MediaTypes) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

// UnmarshalText deserializes string data.
func (l *MediaTypes) UnmarshalText(data []byte) (err error) {
	*l, err = ParseListStrict(string(data))
	return
}

// Scan implements sql.Scanner so MediaType values can be read from databases transparently.
// Currently, only database types that map to string are supported.
// Please consult database-specific driver documentation for matching types.
func (t *MediaType) Scan(value any) error {
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("unable to scan type %T into MediaType", value)
	}
	m, err := Parse(s)
	if err != nil {
		return err
	}
	*t = m
	return nil
}

// Value implements sql.Valuer so that MediaType values can be written to databases transparently.
// Currently, MediaType values map to strings.
// Please consult database-specific driver documentation for matching types.
func (t MediaType) Value() (driver.Value, error) {
	return t.String(), nil
}

// Scan implements sql.Scanner so MediaTypes can be read from databases transparently.
// Currently, only database types that map to string are supported.
// Please consult database-specific driver documentation for matching types.
func (l *MediaTypes) Scan(value any) error {
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("unable to scan type %T into MediaType", value)
	}
	parsed, err := ParseListStrict(s)
	if err != nil {
		return err
	}
	*l = parsed
	return nil
}

// Value implements sql.Valuer so that MediaTypes can be written to databases transparently.
// Currently, MediaTypes map to strings.
// Please consult database-specific driver documentation for matching types.
func (l MediaTypes) Value() (driver.Value, error) {
	return l.String(), nil
}
