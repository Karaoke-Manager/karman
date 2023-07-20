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
		song, err := c.svc.GetSong(r.Context(), id)
		if err != nil {
			// TODO: Maybe support 409 for soft deleted?
			_ = render.Render(w, r, apierror.DBError(err))
			return
		}
		ctx := SetSong(r.Context(), song)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
