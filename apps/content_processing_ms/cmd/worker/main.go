package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ecbDeveloper/netflix-architecture/apps/content_processing_ms/internal/config"
	"github.com/ecbDeveloper/netflix-architecture/apps/content_processing_ms/internal/database"
	"github.com/ecbDeveloper/netflix-architecture/apps/content_processing_ms/internal/processor"
	"github.com/ecbDeveloper/netflix-architecture/apps/content_processing_ms/internal/storage"
	"github.com/ecbDeveloper/netflix-architecture/apps/content_processing_ms/internal/transcoder"
	"github.com/rabbitmq/amqp091-go"
)

func main() {
	loggerHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(loggerHandler)

	cfg := config.LoadConfig(logger)

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	repo, err := database.NewRepository(ctx, cfg)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer repo.Close()

	s3Client, err := storage.NewS3Client(ctx, cfg)
	if err != nil {
		slog.Error("Failed to initialize S3 client", "error", err)
		os.Exit(1)
	}

	transcoderService := transcoder.NewTranscoder()

	connStr := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.RabbitMQUser, cfg.RabbitMQPass, cfg.RabbitMQHost, cfg.RabbitMQPort)
	rabbitMQConn, err := amqp091.Dial(connStr)
	if err != nil {
		slog.Error("Failed to connect to RabbitMQ", "error", err)
		os.Exit(1)
	}
	defer rabbitMQConn.Close()

	rabbitMQCh, err := rabbitMQConn.Channel()
	if err != nil {
		slog.Error("Failed to open RabbitMQ channel", "error", err)
		os.Exit(1)
	}
	defer rabbitMQCh.Close()

	_, err = rabbitMQCh.QueueDeclare(
		cfg.ContentQueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		slog.Error("Failed to declare queue", "error", err)
		os.Exit(1)
	}

	worker := processor.NewWorker(repo, s3Client, transcoderService, rabbitMQCh, cfg.ContentQueueName, logger)

	logger.Info("Starting worker...")
	worker.Start(ctx)
	logger.Info("Worker stoped gracefully")
}
