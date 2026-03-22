package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// Recovery recovers from panics and returns a structured error response.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered",
					"error", err,
					"stack", string(debug.Stack()),
					"path", r.URL.Path,
					"method", r.Method,
				)
				appErr := apperror.NewInternal(fmt.Errorf("panic: %v", err))
				apperror.WriteError(w, appErr)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
