package middleware

// contextKey is the base type used for context keys defined in this package.
// This type is intentionally private.
// Instead, context values should be accessed via the appropriate accessor functions.
type contextKey int

const (
	// contextKeyPagination is a context for that stores a Pagination value.
	contextKeyPagination contextKey = iota
)
