package http

import (
	"context"
	"net/http"

	"codeberg.org/manuelarte/logevent"
	"codeberg.org/manuelarte/logevent/internal"
	"codeberg.org/manuelarte/logevent/middlewares"
)

// AddLogEventMiddleware returns a middleware that adds a request-scoped logger to the context.
//
// Parameters:
//   - t: The log event struct template that will be copied for each request and logged
//     after the request is served. Must be a type whose pointer implements logevent.LogEvent.
//   - f: A function that returns the logging framework to use. It's called when logging,
//     allowing you to provide a context-aware logger.
//
// The middleware wraps each request, creates a per-request copy of the log event, stores it
// in the context, and logs it after the handler completes. Handlers can update the log event
// using UpdateLogEvent during request processing.
//
// Example:
//
//	type RequestLog struct {
//		Method   string
//		Path     string
//		Status   int
//		Duration time.Duration
//	}
//
//	func (l RequestLog) Log(ctx context.Context, li logevent.LogInterface) {
//		li.Info("request completed", slog.String("method", l.Method), slog.Int("status", l.Status))
//	}
//
//	handler := AddLogEventMiddleware(RequestLog{}, func(ctx context.Context) logevent.LogInterface {
//		return slog.Default()
//	})(nextHandler)
func AddLogEventMiddleware[T any, PT internal.PtrLogEvent[T]](
	t T,
	f func(ctx context.Context) logevent.LogInterface,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewares.HandleWithLogEvent[T, PT](r.Context(), t, f, func(ctx context.Context) {
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
			})
		})
	}
}
