package uploads

import (
	"context"
	"github.com/Karaoke-Manager/karman/internal/api/apierror"
	"github.com/Karaoke-Manager/karman/internal/model"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"github.com/go-chi/chi/v5"
	"io/fs"
	"net/http"
)

type contextKey int

const (
	contextKeyFilePath contextKey = iota
	contextKeyUploadInstance
)

func SetUpload(ctx context.Context, upload model.Upload) context.Context {
	return context.WithValue(ctx, contextKeyUploadInstance, upload)
}

func GetUpload(ctx context.Context) (model.Upload, bool) {
	u, ok := ctx.Value(contextKeyUploadInstance).(model.Upload)
	return u, ok
}

func MustGetUpload(ctx context.Context) model.Upload {
	return ctx.Value(contextKeyUploadInstance).(model.Upload)
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
			_ = apierror.InvalidUploadPath(w, r, path)
			return
		}
		ctx := SetFilePath(r.Context(), path)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
