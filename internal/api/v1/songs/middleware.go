package songs

import (
	"context"
	"github.com/Karaoke-Manager/karman/internal/api/apierror"
	"github.com/Karaoke-Manager/karman/internal/api/middleware"
	"github.com/Karaoke-Manager/karman/internal/model"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"net/http"
)

// contextKey is the type for context keys used in this package.
// This type is intentionally private.
// Use the accessor functions instead to interact with context values.
type contextKey int

const (
	// contextKeyInstance identifies a Song instance in a context.
	contextKeyInstance contextKey = iota
)

// SetSong sets the song instance in ctx.
func SetSong(ctx context.Context, song model.Song) context.Context {
	return context.WithValue(ctx, contextKeyInstance, song)
}

// GetSong returns a model.Song instance from the context.
// If the context does not contain a song instance, the second return value will be false.
func GetSong(ctx context.Context) (model.Song, bool) {
	u, ok := ctx.Value(contextKeyInstance).(model.Song)
	return u, ok
}

// MustGetSong returns a model.Song instance from the context.
// In contrast to GetSong this function panics if the context does not contain a song instance.
func MustGetSong(ctx context.Context) model.Song {
	return ctx.Value(contextKeyInstance).(model.Song)
}

// FetchUpload is a middleware that fetches the model.Song instance identified by the request and stores it in the request context.
func (c Controller) FetchUpload(includeFiles bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			id := middleware.MustGetUUID(r.Context())
			var song model.Song
			var err error
			if includeFiles {
				song, err = c.svc.GetSongWithFiles(r.Context(), id)
			} else {
				song, err = c.svc.GetSong(r.Context(), id)
			}
			if err != nil {
				// TODO: Maybe support 410 for soft deleted?
				_ = render.Render(w, r, apierror.DBError(err))
				return
			}
			ctx := SetSong(r.Context(), song)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

// CheckModify is a middleware that checks if modifications to the requested resource are allowed.
func (c Controller) CheckModify(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		song := MustGetSong(r.Context())
		if song.UploadID != nil {
			_ = render.Render(w, r, apierror.UploadSongReadonly(song))
			return
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}