// Package logevent provides utilities to implement canonical logging (wide log events) in Go.
//
// Canonical logging emits a single structured log entry at the end of a unit of work (e.g., HTTP request,
// gRPC call, or background task) capturing the full lifecycle context, metadata, and status.
//
// The package provides core context helpers ([AddLogEventToContext], [UpdateLogEvent], [LogItFunc], [HandleWithLogEvent])
// as well as ready-to-use middleware for HTTP (in subpackage mw/http) and gRPC (in subpackage mw/grpc).
package logevent
