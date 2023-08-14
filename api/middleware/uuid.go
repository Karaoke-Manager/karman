package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/pkg/render"
)

// UUID is a simple middleware that fetches a UUID value from a request parameter named param.
// The UUID will be stored in the request context and can be fetched via GetUUID.
func UUID(param string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, param)
			v, err := uuid.Parse(id)
			if err != nil {
				_ = render.Render(w, r, apierror.ErrInvalidUUID)
				return
			}
			ctx := SetUUID(r.Context(), v)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

// SetUUID sets u in ctx.
// The value is retrievable later via GetUUID and MustGetUUID.
func SetUUID(ctx context.Context, u uuid.UUID) context.Context {
	return context.WithValue(ctx, contextKeyUUID, u)
}

// GetUUID returns the UUID value from the request, if any.
// If the request does not contain a UUID value, the second return value will be false.
func GetUUID(ctx context.Context) (p uuid.UUID, ok bool) {
	p, ok = ctx.Value(contextKeyUUID).(uuid.UUID)
	return
}

// MustGetUUID returns the UUID value from the request.
// In contrast to GetUUID this function panics if the request does not contain a UUID value.
func MustGetUUID(ctx context.Context) uuid.UUID {
	return ctx.Value(contextKeyUUID).(uuid.UUID)
}
