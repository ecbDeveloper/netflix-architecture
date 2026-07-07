package main

import (
	"context"
	"encoding/gob"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/config"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/server"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/shared"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
)

func init() {
	gob.Register(uuid.UUID{})
}

func main() {
	jsonHandler := slog.NewJSONHandler(os.Stdout, nil)
	otelHandler := otelslog.NewHandler(shared.ServerName)

	multiHandler := slogmulti.Fanout(otelHandler, jsonHandler)

	loggerHandler := config.NewLogContextHandler(multiHandler)
	logger := slog.New(loggerHandler)

	if err := godotenv.Load(); err != nil {
		logger.Warn("failed to load .env file", slog.Any("error", err))
	}

	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid configuration", slog.Any("error", err))
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	server.Run(ctx, logger, cfg)
}
