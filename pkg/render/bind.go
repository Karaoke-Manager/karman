package render

import (
	"fmt"
	"net/http"
	"reflect"
)

// Binder interface for managing request payloads.
// Implementing this interface allows a value to be used in the [Bind] function.
// Binding happens after decoding is finished and gives you the opportunity to do
// as much post-processing as necessary.
//
// Implementing this interface also gives way to perform schema validations.
type Binder interface {
	// The Bind method is invoked when an object is bound to request data.
	// Implementations should perform schema validations that cannot be expressed otherwise
	// and potentially perform data normalization.
	//
	// The error returned by this method is returned (as a wrapped error) to the caller of the [Bind] function.
	Bind(r *http.Request) error
}

// NopBinder implements the Binder interface.
// The implementation does nothing.
type NopBinder struct{}

// Bind is an empty implementation of the Binder interface.
func (NopBinder) Bind(*http.Request) error { return nil }

// A DecodeError indicates that the request data could not be properly decoded into the desired data format.
// This could be because of a syntax error or a schema validation error.
// You may inspect the underlying error for more details.
type DecodeError struct {
	err error // underlying error
}

// Error implements the error interface.
func (d DecodeError) Error() string {
	return fmt.Sprintf("decode error: %s", d.err.Error())
}

// Unwrap implements the error interface.
func (d DecodeError) Unwrap() error {
	return d.err
}

// A BindError indicates that the request could be successfully parsed but could not be bound to the model instance.
// This is usually an indication that some kind of constraint imposed by the model's [Binder.Bind] method was violated.
type BindError struct {
	// The underlying error.
	err error
}

// Error implements the error interface.
func (b BindError) Error() string {
	return fmt.Sprintf("bind error: %s", b.err.Error())
}

// Unwrap implements the error interface.
func (b BindError) Unwrap() error {
	return b.err
}

// Bind decodes a request via the [Decode] function and executes the v.Bind method.
// If v is a struct, map or slice type and any fields, slice or map values implement the [Binder] interface
// the Bind methods of those values are called recursively in a bottom-up fashion.
// Note however that the recursion stops at the first value that does not implement the [Binder] interface.
// If you do not need to implement [Binder] yourself but want to give your struct fields the opportunity to perform
// their Bind operations, use the [NopBinder] type to add a noop [Binder] implementation to your type.
//
// For details on the decoding process see the [Decode] function.
//
// There are two main error types returned by this function:
//   - A [DecodeError] indicates an error during the decoding phase.
//     This usually corresponds to a 400 status code.
//   - A [BindError] indicates an error during the invocation of a Bind method
//     (not necessarily v itself but maybe one of its struct fields).
//     This usually corresponds to a 422 status code.
func Bind(r *http.Request, v Binder) error {
	if err := Decode(r, v); err != nil {
		return DecodeError{err}
	}
	if err := binder(r, v); err != nil {
		return BindError{err}
	}
	return nil
}

// binder invokes the Binder.Bind function on v as well as all struct
// fields of v. If v contains fields that implement the Binder interface the
// Binder.Bind function is invoked for those fields in a bottom-up fashion,
// that is v.Bind is invoked last.
func binder(r *http.Request, v Binder) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	// Call Binder on non-struct types right away
	if rv.Kind() != reflect.Struct {
		return v.Bind(r)
	}

	// For structs, we call Bind on each field that implements Binder
	for i := 0; i < rv.NumField(); i++ {
		f := rv.Field(i)
		if isNil(f) {
			continue
		}
		switch f.Type().Kind() {
		case reflect.Slice:
			if f.Type().Elem().Implements(binderType) {
				for i := 0; i < f.Len(); i++ {
					if err := binder(r, f.Index(i).Interface().(Binder)); err != nil {
						return err
					}
				}
			}
		case reflect.Map:
			if f.Type().Elem().Implements(binderType) {
				i := f.MapRange()
				for i.Next() {
					mv := i.Value()
					if err := binder(r, mv.Interface().(Binder)); err != nil {
						return err
					}
				}
			}
		}
		if f.Type().Implements(binderType) {
			fv := f.Interface().(Binder)
			if err := binder(r, fv); err != nil {
				return err
			}
		}
	}

	// We call it bottom-up
	if err := v.Bind(r); err != nil {
		return err
	}

	return nil
}

var binderType = reflect.TypeOf(new(Binder)).Elem()
