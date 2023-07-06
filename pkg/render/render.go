package render

import (
	"net/http"
	"reflect"
)

// Renderer interface for managing response payloads.
type Renderer interface {
	Render(w http.ResponseWriter, r *http.Request) error
}

// Binder interface for managing request payloads.
type Binder interface {
	Bind(r *http.Request) error
}

// Bind decodes a request body and executes the Binder method of the
// payload structure.
func Bind(r *http.Request, v Binder) error {
	if err := Decode(r, v); err != nil {
		return err
	}
	return binder(r, v)
}

// Render renders a single payload and respond to the client request.
func Render(w http.ResponseWriter, r *http.Request, v Renderer) error {
	if err := renderer(w, r, v); err != nil {
		return err
	}
	return Respond(w, r, v)
}

// RenderList renders a slice of payloads and responds to the client request.
func RenderList(w http.ResponseWriter, r *http.Request, l []Renderer) error {
	for _, v := range l {
		if err := renderer(w, r, v); err != nil {
			return err
		}
	}
	return Respond(w, r, l)
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
// Renderer.Render function is invoked for those fields in a bottom-down fashion,
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
		if f.Type().Implements(rendererType) {

			if isNil(f) {
				continue
			}

			fv := f.Interface().(Renderer)
			if err := renderer(w, r, fv); err != nil {
				return err
			}

		}
	}

	// We call it bottom-down.
	if err := v.Render(w, r); err != nil {
		return err
	}

	return nil
}

// binder invokes the Binder.Bind function on v as well as all struct
// fields of v. If v contains fields that implement the Binder interface the
// Binder.Bind function is invoked for those fields in a bottom-down fashion,
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
		if f.Type().Implements(binderType) {

			if isNil(f) {
				continue
			}

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
