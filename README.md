# LogEvent

This library provides utilities to implement the concept of emitting one wide log event
after processing a request in Go.

The steps are the following:

- We define a struct that we are going to update/populate when serving a request.
- We implement the [`Log`](./model.go#L9) method of the [`LogEvent`](./model.go) interface.
- When serving the request, we populate that struct with all the useful information that we want to see in a log entry.
- Once the request is served, the library will log that wide event by calling the method `Log` we implemented.

This is better described in [loggingsucks][loggingsucks].

To see it directly in action, check the [examples](examples) folder.

## Requirements

- Go `1.25.0` or newer

## ⬇️ How to get it

```bash
go get github.com/manuelarte/logevent
```

🚀 ## Features

The library provides a generic function that can be used
to implement the concept of adding a LogEvent to a `context.Context`, then do
some work, and then `Log` that `LogEvent`.

But it also provides some out-of-the-box implementations for:

### HTTP Middleware

This library provides a middleware that can be used to emit a log event after an HTTP request.

```go
// Step 1. Define your log event struct and how to log it.
type transferLogEvent struct {
  Source string
  Target string
  Amount string
  Err    error
}

// Log the event either with Info if everything succeeded or with Error if there was an error.
func (e transferLogEvent) Log(ctx context.Context, li *slog.Logger) {
  if e.Err != nil {
    li.ErrorContext(
      ctx,
      "Error when transferring money",
      slog.String("source", e.Source),
      slog.String("target", e.Target),
      slog.String("amount", e.Amount),
      slog.Any("error", e.Err),
    )
    return
  }

  li.InfoContext(
    ctx,
    "Money transferred successfully",
    slog.String("source", e.Source),
    slog.String("target", e.Target),
    slog.String("amount", e.Amount),
  )
}

// Step 2. Add the middleware to your endpoint.
func registerRoutes() {
  http.Handle(
    "/my-endpoint",
    logeventmiddleware.AddLogEventMiddleware(transferLogEvent{}, func(_ context.Context) *slog.Logger {
      return slog.Default()
    })(http.HandlerFunc(myHandler)),
  )
}

func myHandler(w http.ResponseWriter, r *http.Request) {
  // Step 3. Update your log event while serving the request.
  _ = logeventmiddleware.UpdateLogEvent(r.Context(), func(t *transferLogEvent) {
    t.Source = "Alice"
    t.Target = "Bob"
    t.Amount = "100"
  })
  ...
  err := transferMoney("Alice", "Bob", 100)
  _ = logeventmiddleware.UpdateLogEvent(r.Context(), func(t *transferLogEvent) {
    t.Err = err
  })
  ...
}
```

### gRPC Interceptor

This library also provides a unary server interceptor for your gRPC server.

```go
// Step 1. Define your log event struct and how to log it.
type transferLogEvent struct {
  Source string
  Target string
  Amount string
  Err    error
}

// Log the event either with Info if everything succeeded or with Error if there was an error.
func (e transferLogEvent) Log(ctx context.Context, li *slog.Logger) {
  if e.Err != nil {
    li.ErrorContext(
      ctx,
      "Error when transferring money",
      slog.String("source", e.Source),
      slog.String("target", e.Target),
      slog.String("amount", e.Amount),
      slog.Any("error", e.Err),
    )
    return
  }

  li.InfoContext(
    ctx,
    "Money transferred successfully",
    slog.String("source", e.Source),
    slog.String("target", e.Target),
    slog.String("amount", e.Amount),
  )
}

// Step 2. Add the interceptor to your server.
server := grpc.NewServer(
  grpc.UnaryInterceptor(
    logeventgrpc.UnaryServerInterceptor(transferLogEvent{}, slog.Default()),
  ),
)

func (s transferMoneyServer) Transfer(ctx context.Context, req *TransferMoneyRequest) (*TransferMoneyResponse, error) {
  // Step 3. Update your log event while handling the request.
  _ = logeventmiddleware.UpdateLogEvent(ctx, func(t *transferLogEvent) {
    t.Source = "Alice"
    t.Target = "Bob"
    t.Amount = "100"
  })
  ...
  err := transferMoney("Alice", "Bob", 100)
  _ = logeventmiddleware.UpdateLogEvent(ctx, func(t *transferLogEvent) {
    t.Err = err
  })
  ...
}
```

## Architecture

This library provides an HTTP middleware and a gRPC interceptor, but also a
[generic implementation](./middlewares/middleware.go) for a custom way to serve a request that encapsulates:

1. Creating a per-request copy of the log event struct
2. Wrapping it with thread-safe access (concurrency support)
3. Storing it in the request context
4. Deferring the log output until after the request handler completes
5. Checking for any updates made by the handler

This ensures consistent behavior and makes it easy to update the logging logic in a single place.

## Examples

For runnable examples check the [examples](examples) folder.

- HTTP example: [`examples/http/main.go`](examples/http/main.go)
- gRPC example: [`examples/grpc/main.go`](examples/grpc/main.go)

[loggingsucks]: https://loggingsucks.com
