// Package apierror provides two things: an implementation of [RFC 9457] and
// convenience functions that generate the errors returned by the Karman API.
//
// The ProblemDetails provides an implementation of [RFC 9457].
// You can use it to provide as many or as few details about an error as necessary.
// This package is intended to be used together with render.Render but can be used on its own as well.
//
// [RFC 9457]: https://datatracker.ietf.org/doc/html/rfc9457
package apierror
