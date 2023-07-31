package render

import (
	"fmt"
	"net/http"
	"reflect"
)

// Renderer interface for managing response payloads.
// Implementing this interface allows a value to be used in the [Render] function.
// Rendering occurs before a response is sent, allowing you to perform normalization and cleanup.
//
// Implementing this interface also gives way to ensure schema compliance.
type Renderer interface {
	// The Render method is invoked when an object is about to be rendered (aka serialized).
	// Implementing this method gives you the opportunity to perform data sanitization and normalization
	// before an object is marshalled.
	//
	// Any errors returned by this method will be wrapped and passed on to the caller of the [Render] function.
	// Note however that this late in the request lifecycle there are usually few options of actually handling such an error.
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

// Render executes v.Render and then encodes v using the [Respond] function.
// If v is a struct, map or slice type and any of its fields implement the [Renderer] interface,
// the [Renderer.Render] methods will be called recursively in a bottom-up fashion.
// Note however that the recursion stops at the first value that does not implement the [Renderer] interface.
// If you do not need to implement [Renderer] yourself but want to give your struct fields the opportunity to perform
// their [Renderer.Render] operations, use the [NopRenderer] type to add a noop [Renderer] implementation to your type.
//
// For details on the encoding process see the [Respond] function.
//
// There are two main error types returned by this function:
//   - A [RenderError] indicates an error during the rendering phase.
//     This means that one of the [Renderer.Render] implementations has returned an error.
//   - A [RespondError] indicates an error during the invocation of the [Respond] function.
//     This is usually an indication of a network error.
func Render(w http.ResponseWriter, r *http.Request, v Renderer) error {
	if err := renderer(w, r, v); err != nil {
		return RenderError{err}
	}
	if err := Respond(w, r, v); err != nil {
		return RespondError{err}
	}
	return nil
}

// RenderList works like [Render] but takes a slice of payloads.
// See [Render] for details.
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

var rendererType = reflect.TypeOf(new(Renderer)).Elem()
