package grpc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/manuelarte/logevent/internal"
	"github.com/manuelarte/logevent/middlewares"
)

// UnaryServerInterceptor returns a new unary server interceptor that emits a log event
// once the request is handled.
//
// Parameters:
//   - t: The log event struct template that will be copied for each RPC call and logged
//     after the RPC completes. Must be a type whose pointer implements logevent.LogEvent.
//   - f: A function that returns the logging framework to use. It's called when logging,
//     allowing you to provide a context-aware logger.
//
// The interceptor wraps each RPC call, creates a per-call copy of the log event, stores it
// in the context, and logs it after the handler completes. RPC handlers can update the log event
// using middlewares.UpdateLogEvent during RPC processing.
//
// Example:
//
//	type RPCLog struct {
//		Method   string
//		Error    error
//		Duration time.Duration
//	}
//
//	func (l RPCLog) Log(ctx context.Context, li logevent.LogInterface) {
//		if l.Error != nil {
//			li.Error("RPC failed", slog.String("method", l.Method), slog.Any("error", l.Error))
//		} else {
//			li.Info("RPC succeeded", slog.String("method", l.Method))
//		}
//	}
//
//	interceptor := UnaryServerInterceptor(RPCLog{}, func(ctx context.Context) logevent.LogInterface {
//		return slog.Default()
//	})
//	grpc.NewServer(grpc.ChainUnaryInterceptor(interceptor))
func UnaryServerInterceptor[L, T any, PT internal.PtrLogEvent[L, T]](
	t T,
	f func(ctx context.Context) L,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		var (
			resp any
			err  error
		)

		middlewares.HandleWithLogEvent[L, T, PT](ctx, t, f, func(ctxWithLogEvent context.Context) {
			resp, err = handler(ctxWithLogEvent, req)
		})

		return resp, err
	}
}
