package uploads

import (
	"context"
	"github.com/Karaoke-Manager/karman/internal/api/apierror"
	"github.com/Karaoke-Manager/karman/internal/entity"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"github.com/go-chi/chi/v5"
	"io/fs"
	"net/http"
)

type contextKey int

const (
	contextKeyFilePath contextKey = iota
	contextKeyInstance
)

func SetUpload(ctx context.Context, upload entity.Upload) context.Context {
	return context.WithValue(ctx, contextKeyInstance, upload)
}

func GetUpload(ctx context.Context) (entity.Upload, bool) {
	u, ok := ctx.Value(contextKeyInstance).(entity.Upload)
	return u, ok
}

func MustGetUpload(ctx context.Context) entity.Upload {
	return ctx.Value(contextKeyInstance).(entity.Upload)
}

func (c *Controller) fetchUpload(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "uuid")
		upload, err := c.Service.GetUpload(r.Context(), id)
		if err != nil {
			// TODO: Differentiate errors (especially 404)
			// 	     Maybe also 409 Gone for soft delete
			_ = render.Render(w, r, apierror.ErrInternalServerError)
			return
		}
		ctx := SetUpload(r.Context(), upload)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

func SetFilePath(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, contextKeyFilePath, path)
}

func GetFilePath(ctx context.Context) (string, bool) {
	path, ok := ctx.Value(contextKeyFilePath).(string)
	return path, ok
}

func MustGetFilePath(ctx context.Context) string {
	return ctx.Value(contextKeyFilePath).(string)
}

func (c *Controller) validateFilePath(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		path := chi.URLParam(r, "*")
		if !fs.ValidPath(path) {
			_ = render.Render(w, r, apierror.InvalidUploadPath(path))
			return
		}
		ctx := SetFilePath(r.Context(), path)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
