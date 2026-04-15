package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/ecbDeveloper/netflix-architecture/apps/recomendation_ms/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/apps/recomendation_ms/internal/recommendation"
	historypb "github.com/ecbDeveloper/netflix-architecture/proto/history"
	pb "github.com/ecbDeveloper/netflix-architecture/proto/recommendation"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx := context.Background()

	loggerHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(loggerHandler)

	err := godotenv.Load()
	if err != nil {
		logger.Error("failed to load .env file", slog.Any("error", err))
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50052"
	}

	pool, err := initializeDatabaseConnection(ctx)
	if err != nil {
		logger.Error("failed to initialize db pool", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()

	historyAddr := os.Getenv("HISTORY_GRPC_ADDR")
	historyConn, err := grpc.NewClient(historyAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("failed to connect to history ms", slog.Any("error", err))
		os.Exit(1)
	}
	defer historyConn.Close()

	historyClient := historypb.NewHistoryServiceClient(historyConn)

	queries := sqlc.New(pool)
	server := recommendation.NewServer(queries, historyClient)

	grpcServer := grpc.NewServer()
	pb.RegisterRecommendationServiceServer(grpcServer, server)

	if os.Getenv("ENV") == "development" {
		reflection.Register(grpcServer)
	}

	listener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		logger.Error("failed to listen", slog.Any("error", err))
		os.Exit(1)
	}

	logger.Info("recommendation microservice started", slog.String("port", grpcPort))
	if err := grpcServer.Serve(listener); err != nil {
		logger.Error("failed to serve", slog.Any("error", err))
		os.Exit(1)
	}
}

func initializeDatabaseConnection(ctx context.Context) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_NAME"),
	)

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create new db pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return pool, nil
}
