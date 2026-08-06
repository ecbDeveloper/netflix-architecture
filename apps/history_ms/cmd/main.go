package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/ecbDeveloper/netflix-architecture/apps/history_ms/internal/config"
	"github.com/ecbDeveloper/netflix-architecture/apps/history_ms/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/apps/history_ms/internal/history"
	historypb "github.com/ecbDeveloper/netflix-architecture/gen/go/history/v1"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	loggerHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(loggerHandler)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load application configs", slog.Any("error", err))
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	pool, err := initializeDatabaseConnection(ctx, cfg)
	if err != nil {
		logger.Error("failed to initialize db pool", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()

	grpcPort := cfg.GRPCPort

	listener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		logger.Error("failed to listen", slog.Any("error", err))
		os.Exit(1)
	}
	defer listener.Close()

	queries := sqlc.New(pool)
	server := history.NewServer(queries)
	grpcServer := grpc.NewServer()
	historypb.RegisterHistoryServiceServer(grpcServer, server)

	if os.Getenv("ENV") == "development" {
		reflection.Register(grpcServer)
	}

	go func() {
		logger.Info("history microservice started", slog.String("port", grpcPort))
		if err := grpcServer.Serve(listener); err != nil {
			logger.Error("failed to serve", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	grpcServer.GracefulStop()
	logger.Info("History grpc server stoped gracefully")
}

func initializeDatabaseConnection(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to create new db pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return pool, nil
}
