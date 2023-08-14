// Package api contains the HTTP endpoints of the Karman API.
// We use the github.com/go-chi/chi router.
// Every collection of endpoints is implemented as a Controller type
// that exposes a Router function compatible with Chi's Route function.
// Many controllers contain sub-controllers from their respective sub-packages.
//
// Package api also contains some utility packages that are used across API endpoints.
// These include consistent error messages and middlewares used across the API.
package api
