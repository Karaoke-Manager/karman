package uploads

import (
	"context"
	"io/fs"
	"net/http"
	"slices"

	"github.com/go-chi/chi/v5"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/api/middleware"
	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/render"
)

type contextKey int

const (
	contextKeyInstance contextKey = iota
	contextKeyFilePath
)

// SetUpload sets the song instance in ctx.
func SetUpload(ctx context.Context, upload model.Upload) context.Context {
	return context.WithValue(ctx, contextKeyInstance, upload)
}

// GetUpload returns a model.Upload instance from the context.
// If the context does not contain an upload instance, the second return value will be false.
func GetUpload(ctx context.Context) (model.Upload, bool) {
	u, ok := ctx.Value(contextKeyInstance).(model.Upload)
	return u, ok
}

// MustGetUpload returns a model.Upload instance from the context.
// In contrast to GetUpload this function panics if the context does not contain an upload instance.
func MustGetUpload(ctx context.Context) model.Upload {
	return ctx.Value(contextKeyInstance).(model.Upload)
}

// FetchUpload is a middleware that fetches the model.Upload instance identified by the request and stores it in the request context.
func (c *Controller) FetchUpload(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		id := middleware.MustGetUUID(r.Context())
		// TODO: Maybe support 410 for soft deleted?
		upload, err := c.uploadRepo.GetUpload(r.Context(), id)
		if err != nil {
			_ = render.Render(w, r, apierror.ServiceError(err))
			return
		}
		ctx := SetUpload(r.Context(), upload)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// SetFilePath sets the file path in ctx.
func SetFilePath(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, contextKeyFilePath, path)
}

// GetFilePath returns the file path from the context.
// If the context does not contain a file path, the second return value will be false.
func GetFilePath(ctx context.Context) (string, bool) {
	path, ok := ctx.Value(contextKeyFilePath).(string)
	return path, ok
}

// MustGetFilePath returns the file path from the context.
// In contrast to GetFilePath this function panics if the context does not contain a file path.
func MustGetFilePath(ctx context.Context) string {
	return ctx.Value(contextKeyFilePath).(string)
}

// ValidateFilePath is a middleware that validates the file path within an upload syntactically.
// If this middleware passes it is not guaranteed that the path actually exists.
// It is also not guaranteed that the path is allowed for a specific endpoint.
//
// The root directory is normalized to ".".
func ValidateFilePath(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		path := chi.URLParam(r, "*")
		if path == "" {
			path = "."
		}
		if !fs.ValidPath(path) {
			_ = render.Render(w, r, apierror.InvalidUploadPath(path))
			return
		}
		ctx := SetFilePath(r.Context(), path)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// UploadState is a middleware that checks if the upload is in one of the allowed states.
func UploadState(states ...model.UploadState) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			upload := MustGetUpload(r.Context())
			if !slices.Contains(states, upload.State) {
				_ = render.Render(w, r, apierror.UploadState(upload))
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
