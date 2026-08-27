package logevent

import (
	"context"
)

type (
	// logEventKey represents the key to be used to store the LogEvent in the context.
	// It is a generic type that ensures type-safety when storing and retrieving log events.
	logEventKey[L Logger, T any, PT PtrLogEvent[L, T]] struct{}
)

// HandleWithLogEvent is a generic helper function that encapsulates the common pattern of adding a log event
// to the context and executing a handler. It is used by both the HTTP middleware and gRPC interceptor.
//
// The function performs the following steps:
//  1. Creates a per-request copy of the log event struct (to avoid concurrent modifications)
//  2. Type-asserts the copy to get the pointer type (required by the constraint)
//  3. Creates a wrapper around the pointer for concurrency support ([sync.Once], [sync.RWMutex])
//  4. Stores the wrapper in the context under a type-safe key
//  5. Defers a call to log the event after the handler completes
//  6. Calls the provided handler function with the updated context
//  7. Checks if the handler updated the log event in the context and uses the updated version
//
// This design allows handlers to update the log event during request processing and ensures
// the log event is only logged once and is thread-safe.
func HandleWithLogEvent[L Logger, T any, PT PtrLogEvent[L, T]](
	ctx context.Context,
	t T,
	logger L,
	handler func(context.Context),
) {
	tCopy := t // per-request copy

	ctx, deferFunc := AddLogEventToContext[L, T, PT](ctx, tCopy)
	defer deferFunc(logger)

	handler(ctx)
}

// AddLogEventToContext adds a log event to the context.
// It can be used to custom add a log event to the context in any kind of scenario.
//
// The function performs the following steps:
//  1. Type-asserts the provided log event to get the pointer type (required by the constraint)
//  2. Creates a wrapper around the pointer for concurrency support ([sync.Once], [sync.RWMutex])
//  3. Stores the wrapper in the context under a type-safe key
//
// This design allows handlers to update the log event during request processing and ensures
// the log event is only logged once and is thread-safe.
// To add more context to the log event, use UpdateLogEvent.
func AddLogEventToContext[L Logger, T any, PT PtrLogEvent[L, T]](
	parent context.Context,
	t T,
) (context.Context, func(l L)) {
	// Hack: bridge *T -> PT through interface assertion.
	pt, ok := any(&t).(PT)
	if !ok {
		panic("invalid type arguments: expected PT to be *T implementing logevent.LogEvent")
	}

	wle := &wrapperLogEvent[L, T, PT]{
		le: pt,
	}
	newCtx := context.WithValue(parent, logEventKey[L, T, PT]{}, wle)

	return newCtx, func(l L) { wle.Log(newCtx, l) }
}

// UpdateLogEvent updates the log event stored in the context during request processing.
// It works with HTTP middleware, gRPC interceptors, or manual log event context lifecycle,
// allowing handlers to modify the log event that will be logged after the unit of work completes.
//
// Parameters:
//   - ctx: The context containing the log event.
//   - f: A function that receives the pointer to the log event struct and mutates it.
//
// Returns an error if the log event was not initialized (i.e., the unit of work was not wrapped
// with AddLogEventMiddleware, UnaryServerInterceptor, or AddLogEventToContext).
//
// Example with HTTP:
//
//	func myHandler(w http.ResponseWriter, r *http.Request) {
//		_ = logevent.UpdateLogEvent(r.Context(), func(log *RequestLog) {
//			log.Path = r.URL.Path
//			log.Method = r.Method
//		})
//	}
//
// Example with gRPC:
//
//	func (s *server) MyRPC(ctx context.Context, req *pb.Request) (*pb.Response, error) {
//		_ = logevent.UpdateLogEvent(ctx, func(log *RPCLog) {
//			log.Method = "MyRPC"
//		})
//		return &pb.Response{}, nil
//	}
func UpdateLogEvent[L Logger, T any, PT PtrLogEvent[L, T]](ctx context.Context, f func(t PT)) error {
	v, ok := ctx.Value(logEventKey[L, T, PT]{}).(*wrapperLogEvent[L, T, PT])
	if !ok {
		return ErrLogEventNotInitialized
	}

	return v.Update(f)
}
