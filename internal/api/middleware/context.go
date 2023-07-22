package middleware

// contextKey is the base type used for context keys defined in this package.
// This type is intentionally private.
// Instead, context values should be accessed via the appropriate accessor functions.
type contextKey int

const (
	// contextKeyPagination is a context key for that stores a Pagination value.
	contextKeyPagination contextKey = iota
	// contextKeyUUID is a context key that stores a UUID value.
	contextKeyUUID
)
