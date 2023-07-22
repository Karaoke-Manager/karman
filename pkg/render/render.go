package render

import (
	"fmt"
	"net/http"
	"reflect"
)

// Renderer interface for managing response payloads.
type Renderer interface {
	// The Render method is invoked when an object is about to be rendered (aka serialized).
	// Implementing this method gives you the opportunity to perform data sanitization and normalization
	// before an object is marshalled.
	Render(w http.ResponseWriter, r *http.Request) error
}

// NopRenderer implements the Renderer interface.
// The implementation does nothing.
type NopRenderer struct{}

// Render is an empty implementation of the Renderer interface.
func (NopRenderer) Render(http.ResponseWriter, *http.Request) error { return nil }

// A RenderError indicates that an error occurred during the rendering process of the response.
// This kind of error is often an indication of a programming error or can be
// an indication of malformed data.
type RenderError struct {
	// The underlying error.
	err error
}

// Error implements the error interface.
func (r RenderError) Error() string {
	return fmt.Sprintf("render error: %s", r.err.Error())
}

// Unwrap implements the error interface.
func (r RenderError) Unwrap() error {
	return r.err
}

// A RespondError indicates that an error occurred during sending the rendered response.
// This is often an indication of a network problem or some kind of system failure outside the control of the program.
type RespondError struct {
	// The underlying error.
	err error
}

// Error implements the error interface.
func (r RespondError) Error() string {
	return fmt.Sprintf("respond error: %s", r.err.Error())
}

// Unwrap implements the error interface.
func (r RespondError) Unwrap() error {
	return r.err
}

// Binder interface for managing request payloads.
type Binder interface {
	// The Bind method is invoked when an object is bound to request data.
	// Implementing this method gives you the opportunity to perform data
	// validation and sanitization of input data.
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
	// The underlying error.
	err error
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
// This is usually an indication that some kind of constraint imposed by the model's Bind method was violated.
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

// Bind decodes a request body and executes the Binder method of the
// payload structure.
func Bind(r *http.Request, v Binder) error {
	if err := Decode(r, v); err != nil {
		return DecodeError{err}
	}
	if err := binder(r, v); err != nil {
		return BindError{err}
	}
	return nil
}

// Render renders a single payload and respond to the client request.
func Render(w http.ResponseWriter, r *http.Request, v Renderer) error {
	if err := renderer(w, r, v); err != nil {
		return RenderError{err}
	}
	if err := Respond(w, r, v); err != nil {
		return RespondError{err}
	}
	return nil
}

// RenderList renders a slice of payloads and responds to the client request.
func RenderList(w http.ResponseWriter, r *http.Request, l []Renderer) error {
	for _, v := range l {
		if err := renderer(w, r, v); err != nil {
			return RenderError{err}
		}
	}
	if err := Respond(w, r, l); err != nil {
		return RespondError{err}
	}
	return nil
}

// isNil is a helper function that tests if f is the nil value.
func isNil(f reflect.Value) bool {
	switch f.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return f.IsNil()
	default:
		return false
	}
}

// renderer invokes the Renderer.Render function on v as well as all struct
// fields of v. If v contains fields that implement the Renderer interface the
// Renderer.Render function is invoked for those fields in a bottom-up fashion,
// that is v.Render is invoked last.
func renderer(w http.ResponseWriter, r *http.Request, v Renderer) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	// We're done if the Renderer isn't a struct object
	if rv.Kind() != reflect.Struct {
		return v.Render(w, r)
	}

	// For structs, we call Render on each field that implements Renderer
	for i := 0; i < rv.NumField(); i++ {
		f := rv.Field(i)
		if isNil(f) {
			continue
		}
		switch f.Type().Kind() {
		case reflect.Slice:
			if f.Type().Elem().Implements(rendererType) {
				for i := 0; i < f.Len(); i++ {
					if err := renderer(w, r, f.Index(i).Interface().(Renderer)); err != nil {
						return err
					}
				}
			}
		case reflect.Map:
			if f.Type().Elem().Implements(rendererType) {
				i := f.MapRange()
				for i.Next() {
					mv := i.Value()
					if err := renderer(w, r, mv.Interface().(Renderer)); err != nil {
						return err
					}
				}
			}
		}
		if f.Type().Implements(rendererType) {
			fv := f.Interface().(Renderer)
			if err := renderer(w, r, fv); err != nil {
				return err
			}
		}
	}

	// We call it bottom-up.
	if err := v.Render(w, r); err != nil {
		return err
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

var (
	rendererType = reflect.TypeOf(new(Renderer)).Elem()
	binderType   = reflect.TypeOf(new(Binder)).Elem()
)
