package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/pkg/render"
)

// Logger returns a middleware that logs every request after it is done.
func Logger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()
			defer func() {
				elapsed := time.Since(t1)
				logger.LogAttrs(r.Context(), slog.LevelInfo-1, "Request completed.",
					slog.String("method", r.Method),
					slog.String("path", r.RequestURI),
					slog.String("addr", r.RemoteAddr),
					slog.Int("status", ww.Status()),
					slog.Duration("elapsed", elapsed),
				)
			}()
			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}

// Recoverer returns a middleware that recovers from panics, logs the panic (and a backtrace),
// and returns an HTTP 500 (Internal Server Error) status if possible.
//
// If printStack is true, the middleware will print the stack trace of a panic to stderr.
// This can be useful in debugging.
func Recoverer(logger *slog.Logger, printTrace bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					if rvr == http.ErrAbortHandler { //nolint:errorlint
						// we don't recover http.ErrAbortHandler so the response
						// to the client is aborted, this should not be logged
						panic(rvr)
					}

					msg := fmt.Sprintf("%v", rvr)
					stack := string(debug.Stack())
					logger.LogAttrs(r.Context(), slog.LevelError, msg,
						slog.String("method", r.Method),
						slog.String("path", r.RequestURI),
						slog.String("addr", r.RemoteAddr),
						slog.String("stack", stack),
					)
					if printTrace {
						fmt.Fprint(os.Stderr, stack)
					}

					_ = render.Render(w, r, apierror.ErrInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
