package config

import (
	"context"
	"log/slog"

	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/middleware"
)

type LogContextHandler struct {
	slog.Handler
}

func (h *LogContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if id := middleware.GetRequestID(ctx); id != "" {
		r.AddAttrs(slog.String("request_id", id))
	}
	return h.Handler.Handle(ctx, r)
}

func NewLogContextHandler(h slog.Handler) slog.Handler {
	return &LogContextHandler{Handler: h}
}
