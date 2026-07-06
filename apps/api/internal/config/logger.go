package config

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

type LogContextHandler struct {
	slog.Handler
}

func (h *LogContextHandler) Handle(ctx context.Context, r slog.Record) error {
	spanCtx := trace.SpanFromContext(ctx).SpanContext()

	if spanCtx.IsValid() {
		r.AddAttrs(
			slog.String("trace_id", spanCtx.TraceID().String()),
			slog.String("span_id", spanCtx.SpanID().String()),
		)
	}

	return h.Handler.Handle(ctx, r)
}

func NewLogContextHandler(h slog.Handler) slog.Handler {
	return &LogContextHandler{Handler: h}
}
