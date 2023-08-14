package middleware

import (
	"context"
	"net/http"
	"strconv"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/pkg/render"
)

const (
	// PaginationLimitKey is the query parameter specifying the pagination limit.
	PaginationLimitKey = "limit"
	// PaginationOffsetKey is the query parameter specifying the pagination offset.
	PaginationOffsetKey = "offset"
)

// Pagination is a simple data object holding information about the current pagination query.
type Pagination struct {
	// The limit originally found in the request.
	// May exceed the maximum limit and may be negative.
	RequestLimit int

	// The sanitized pagination limit.
	// Non-negative and bounded by the maximum of the pagination middleware.
	Limit int

	// The sanitized pagination offset.
	// The Offset is non-negative but otherwise unbounded.
	Offset int
}

// Paginate is a middleware that processes paginated queries and provides sanitized parameters via the request context.
// This middleware implements limit-offset pagination through two query parameters limit and offset.
// The limit defaults to the specified defaultLimit and is bounded by maxLimit.
// The offset defaults to 0 and is unbounded.
// Both limit and offset will be set to 0 if they are negative.
//
// This middleware may return a 400 response of the pagination query cannot be parsed.
// If the pagination query could be parsed the request context will contain the sanitized data.
// You can get the data via GetPagination and MustGetPagination.
func Paginate(defaultLimit int, maxLimit int) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()
			rawLimit := query.Get(PaginationLimitKey)
			limit := defaultLimit
			var err error
			if rawLimit != "" {
				if limit, err = strconv.Atoi(rawLimit); err != nil {
					// TODO: Maybe a specialized error?
					_ = render.Render(w, r, apierror.ErrBadRequest)
					return
				}
			}
			requestLimit := limit
			if limit < 0 {
				limit = 0
			} else if limit > maxLimit {
				limit = maxLimit
			}
			rawOffset := query.Get(PaginationOffsetKey)
			offset := 0
			if rawOffset != "" {
				if offset, err = strconv.Atoi(rawOffset); err != nil {
					_ = render.Render(w, r, apierror.ErrBadRequest)
					return
				}
			}
			if offset < 0 {
				offset = 0
			}
			ctx := SetPagination(r.Context(), Pagination{
				RequestLimit: requestLimit,
				Limit:        limit,
				Offset:       offset,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

// SetPagination sets p in ctx.
// The value is retrievable later via GetPagination and MustGetPagination.
func SetPagination(ctx context.Context, p Pagination) context.Context {
	return context.WithValue(ctx, contextKeyPagination, p)
}

// GetPagination returns the pagination value from the request, if any.
// If the request does not contain a pagination value, the second return value will be false.
func GetPagination(ctx context.Context) (p Pagination, ok bool) {
	p, ok = ctx.Value(contextKeyPagination).(Pagination)
	return
}

// MustGetPagination returns the pagination value from the request.
// In contrast to GetPagination this function panics if the request does not contain a pagination value.
func MustGetPagination(ctx context.Context) Pagination {
	return ctx.Value(contextKeyPagination).(Pagination)
}
