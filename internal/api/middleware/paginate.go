package middleware

import (
	"context"
	"github.com/Karaoke-Manager/karman/internal/api/apierror"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"net/http"
	"strconv"
)

const (
	PaginationLimitKey  = "limit"
	PaginationOffsetKey = "offset"
)

type Pagination struct {
	Limit  int
	Offset int
}

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
			if limit > maxLimit {
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
			ctx := SetPagination(r.Context(), Pagination{
				Limit:  limit,
				Offset: offset,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

func SetPagination(ctx context.Context, p Pagination) context.Context {
	return context.WithValue(ctx, contextKeyPagination, p)
}

func GetPagination(ctx context.Context) (Pagination, bool) {
	p, ok := ctx.Value(contextKeyPagination).(Pagination)
	return p, ok
}

func MustGetPagination(ctx context.Context) Pagination {
	return ctx.Value(contextKeyPagination).(Pagination)
}
