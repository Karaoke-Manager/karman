package songs

import (
	"context"
	"github.com/Karaoke-Manager/karman/internal/api/apierror"
	"github.com/Karaoke-Manager/karman/internal/model"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type contextKey int

const (
	contextKeyInstance contextKey = iota
)

func SetSong(ctx context.Context, song model.Song) context.Context {
	return context.WithValue(ctx, contextKeyInstance, song)
}

func GetSong(ctx context.Context) (model.Song, bool) {
	u, ok := ctx.Value(contextKeyInstance).(model.Song)
	return u, ok
}

func MustGetSong(ctx context.Context) model.Song {
	return ctx.Value(contextKeyInstance).(model.Song)
}

func (c *Controller) fetchUpload(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "uuid")
		song, err := c.Service.GetSong(r.Context(), id)
		if err != nil {
			// TODO: Differentiate errors (404, maybe 409)
			_ = render.Render(w, r, apierror.ErrInternalServerError)
			return
		}
		ctx := SetSong(r.Context(), song)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
