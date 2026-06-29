package middlewares

import (
	"context"

	"github.com/manuelarte/logevent/internal"
)

// HandleWithLogEvent is a generic helper function that encapsulates the common pattern of adding a log event
// to the context and executing a handler. It is used by both the HTTP middleware and gRPC interceptor.
//
// The function performs the following steps:
// 1. Creates a per-request copy of the log event struct (to avoid concurrent modifications)
// 2. Type-asserts the copy to get the pointer type (required by the constraint)
// 3. Creates a wrapper around the pointer for concurrency support (sync.Once, sync.RWMutex)
// 4. Stores the wrapper in the context under a type-safe key
// 5. Defers a call to log the event after the handler completes
// 6. Calls the provided handler function with the updated context
// 7. Checks if the handler updated the log event in the context and uses the updated version
//
// This design allows handlers to update the log event during request processing, and ensures
// the log event is only logged once and is thread-safe.
func HandleWithLogEvent[L, T any, PT internal.PtrLogEvent[L, T]](
	ctx context.Context,
	t T,
	loggerFunc func(context.Context) L,
	handler func(context.Context),
) *internal.WrapperLogEvent[L, T, PT] {
	tCopy := t // per-request copy

	// Hack: bridge *T -> PT through interface assertion.
	pt, ok := any(&tCopy).(PT)
	if !ok {
		panic("invalid type arguments: expected PT to be *T implementing logevent.LogEvent")
	}

	wle := internal.NewWrapperLogEvent(pt)

	defer func(ctx context.Context) {
		wle.Log(ctx, loggerFunc(ctx))
	}(ctx)

	ctx = context.WithValue(ctx, internal.LogEventKey[L, T, PT]{}, wle)
	handler(ctx)

	v, ok := ctx.Value(internal.LogEventKey[L, T, PT]{}).(*internal.WrapperLogEvent[L, T, PT])
	if ok {
		wle = v
	}

	return wle
}

// UpdateLogEvent updates the log event stored in the context during request processing.
// It works with both HTTP middleware and gRPC interceptors, allowing handlers to modify
// the log event that will be logged after the request completes.
//
// Parameters:
//   - ctx: The context containing the log event (from HTTP request or gRPC call)
//   - f: A function that receives the pointer to the log event struct and modifies it
//
// Returns an error if the log event was not initialized (i.e., the request was not wrapped
// with AddLogEventMiddleware or UnaryServerInterceptor).
//
// Example with HTTP:
//
//	func myHandler(w http.ResponseWriter, r *http.Request) {
//		_ = middlewares.UpdateLogEvent(r.Context(), func(log *RequestLog) {
//			log.Path = r.URL.Path
//			log.Method = r.Method
//		})
//	}
//
// Example with gRPC:
//
//	func (s *server) MyRPC(ctx context.Context, req *pb.Request) (*pb.Response, error) {
//		_ = middlewares.UpdateLogEvent(ctx, func(log *RPCLog) {
//			log.Method = "MyRPC"
//		})
//		return &pb.Response{}, nil
//	}
func UpdateLogEvent[L, T any, PT internal.PtrLogEvent[L, T]](ctx context.Context, f func(t PT)) error {
	return internal.UpdateLogEvent[L, T, PT](ctx, f)
}
