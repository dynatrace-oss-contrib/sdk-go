/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// CloudEventTraceContext is a wrapper around the OpenTelemetry TraceContext
// https://github.com/open-telemetry/opentelemetry-go/blob/v1.0.0-RC3/propagation/trace_context.go
type CloudEventTraceContext struct {
	traceContext propagation.TraceContext
}

// NewCloudEventTraceContext creates a new CloudEventTraceContext
func NewCloudEventTraceContext() CloudEventTraceContext {
	return CloudEventTraceContext{traceContext: propagation.TraceContext{}}
}

// Extract extracts the tracecontext from the cloud event into the context.
//
// If the context has a recording span, then the same context is returned. If not, then the extraction
// from the cloud event happens. The auto-instrumentation libraries take care of creating the context
// with the proper/most recent tracecontext, often started by itself. In this case it's more correct
// to take the tracecontext from the context instead of the event.
func (etc CloudEventTraceContext) Extract(ctx context.Context, carrier CloudEventCarrier) context.Context {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		return ctx
	}

	// Extracts the traceparent from the cloud event into a new context
	// This is useful when running code which is not aware of a context (reading from the queue in a long running process)
	// In this case we use the tracecontext inside the event to continue the trace
	return etc.traceContext.Extract(ctx, carrier)
}

// Inject injects the current tracecontext from the context into the cloud event
func (etc CloudEventTraceContext) Inject(ctx context.Context, carrier CloudEventCarrier) {
	etc.traceContext.Inject(ctx, carrier)
}
