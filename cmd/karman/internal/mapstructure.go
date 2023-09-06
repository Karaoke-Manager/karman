package internal

import (
	"encoding"
	"reflect"
)

// TextUnmarshalerDecodeHook is a decoding hook for the mapstructure package.
// The decoding hook implements support for the  encoding.TextUnmarshaler interface.
func TextUnmarshalerDecodeHook(from reflect.Value, to reflect.Value) (any, error) {
	if to.CanAddr() {
		to = to.Addr()
	}
	// If the destination implements the unmarshaling interface
	u, ok := to.Interface().(encoding.TextUnmarshaler)
	if !ok {
		return from.Interface(), nil
	}
	// If it is nil and a pointer, create and assign the target value first
	if to.IsNil() && to.Type().Kind() == reflect.Ptr {
		to.Set(reflect.New(to.Type().Elem()))
		u = to.Interface().(encoding.TextUnmarshaler)
	}
	var text []byte
	switch v := from.Interface().(type) {
	case string:
		text = []byte(v)
	case []byte:
		text = v
	default:
		return v, nil
	}

	if err := u.UnmarshalText(text); err != nil {
		return to.Interface(), err
	}
	return to.Interface(), nil
}
