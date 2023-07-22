package render

// TODO: Maybe publish this as a standalone fork of the CHI repository.

import (
	"context"
	"net/http"
)

// contextKey is a value for use with context.WithValue. It's used as
// a pointer, so it fits in an interface{} without allocation. This technique
// for defining context keys was copied from Go 1.7's new use of context in net/http.
type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "chi render context value " + k.name
}

var (
	// ContextKeyRequestFormat is a context key to record the format of the
	// request payload.
	ContextKeyRequestFormat = &contextKey{"RequestFormat"}

	// ContextKeyResponseFormat is a context key to record the future format of
	// the response payload.
	ContextKeyResponseFormat = &contextKey{"ResponseFormat"}

	// ContextKeyContentType is a context key to record a future content-type.
	ContextKeyContentType = &contextKey{"ContentType"}

	// ContextKeyStatus is a context key to record a future HTTP response status
	// code.
	ContextKeyStatus = &contextKey{"Status"}
)

// Status sets a HTTP response status code hint into request context at any point
// during the request lifecycle. Before the Responder sends its response header
// it will check the ContextKeyStatus.
func Status(r *http.Request, status int) {
	*r = *r.WithContext(context.WithValue(r.Context(), ContextKeyStatus, status))
}

// ContentType sets a response content type hint into the request context at any
// point during the request lifecycle. Before the Responder sends its response
// header it will check the ContextKeyContentType.
func ContentType(r *http.Request, contentType string) {
	*r = *r.WithContext(context.WithValue(r.Context(), ContextKeyContentType, contentType))
}