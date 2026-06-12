package main

import (
	"context"
	"encoding/gob"
	"log/slog"
	"os"

	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/config"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/server"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func init() {
	gob.Register(uuid.UUID{})
}

func main() {
	loggerHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(loggerHandler)

	if err := godotenv.Load(); err != nil {
		logger.Warn("failed to load .env file", slog.Any("error", err))
	}

	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid configuration", slog.Any("error", err))
		os.Exit(1)
	}

	ctx := context.Background()

	server.Run(ctx, logger, cfg)
}
