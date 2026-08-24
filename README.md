# LogEvent

[![CI](https://github.com/manuelarte/logevent/actions/workflows/ci.yml/badge.svg)](https://github.com/manuelarte/logevent/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/manuelarte/logevent.svg)](https://pkg.go.dev/github.com/manuelarte/logevent)

This library provides utilities to implement the concept of emitting one canonical log (wide log event)
after processing a unit of work, inspired by logging patterns from companies like Stripe or Google

This library provides the raw functionality to implement for any unit of work but also provides
two middlewares, one for HTTP and another one for gRPC to be used out of the box.
Check the [examples](./examples) folder for more information.

The steps are the following:

- We define a struct that we are going to update/populate when serving a request.
- We implement the [`Log`](./model.go#L9) method of the [`LogEvent`](./model.go) interface.
This allows us to change the way we want to log the event based on the values.
- When serving the unit of work, we populate that struct event with all the useful information that we want to see in a the log entry.
- Once the unit of work is served, the library will log that canonical log event by calling the method `Log` we implemented.

This is better described in [loggingsucks][loggingsucks].

To see it directly in action, check the [examples](examples) folder.

## Requirements

- Go `1.25.0` or newer

## ⬇️ How to get it

```bash
go get github.com/manuelarte/logevent
```

## 🚀 Features

The library provides a generic function that can be used
to implement the concept of adding a LogEvent to a `context.Context`, then do
some work, and then `Log` that `LogEvent`.

But it also provides some out-of-the-box implementations for:

### Canonical Logging (without middleware)

The library provides primitives to implement canonical logging without needing middleware.
This is useful for background jobs, service layers, or any scenario where you want
to manually control the log event lifecycle.

```go
package main

import (
  "context"
  "log/slog"

  logeventmiddleware "github.com/manuelarte/logevent/mw"
)

type taskLogEvent struct {
  TaskID  string
  Status  string
  Elapsed int64
}

func (e taskLogEvent) Log(ctx context.Context, li *slog.Logger) {
  li.InfoContext(ctx, "Task completed", slog.String("task_id", e.TaskID), slog.String("status", e.Status))
}

func processTask(ctx context.Context, taskID string, logger *slog.Logger) error {
  // Step 1. Add the log event to the context
  ctx = logeventmiddleware.AddLogEventToContext(ctx, taskLogEvent{TaskID: taskID})

  // Step 2. Get the defer function that will log the event
  deferFunc, err := logeventmiddleware.LogEventFunc[*slog.Logger, taskLogEvent, *taskLogEvent](ctx, logger)
  if err != nil {
    return err
  }
  defer deferFunc()

  // Step 3. Update the log event during processing
  _ = logeventmiddleware.UpdateLogEvent(ctx, func(e *taskLogEvent) {
    e.Status = "processing"
  })

  // Do some work...

  // Step 4. Update the log event with final status
  _ = logeventmiddleware.UpdateLogEvent(ctx, func(e *taskLogEvent) {
    e.Status = "completed"
  })

  // The log event is automatically logged when the defer is called
  return nil
}
```

### HTTP Middleware

This library provides a middleware that can be used to emit a log event after an HTTP request.

```go
package main

import (
 "context"
 "log/slog"
 "net/http"

 logeventmiddleware "github.com/manuelarte/logevent/mw"
 logeventhttp "github.com/manuelarte/logevent/mw/http"
)

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
   logeventhttp.AddLogEventMiddleware(transferLogEvent{}, slog.Default())(http.HandlerFunc(myHandler)),
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
package main

import (
 "context"
 "log/slog"

 logeventmiddleware "github.com/manuelarte/logevent/mw"
 logeventgrpc "github.com/manuelarte/logevent/mw/grpc"
)

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

- Canonical logging example: [`examples/canonical/main.go`](examples/canonical/main.go)
- HTTP example: [`examples/http/main.go`](examples/http/main.go)
- gRPC example: [`examples/grpc/main.go`](examples/grpc/main.go)

[loggingsucks]: https://loggingsucks.com
