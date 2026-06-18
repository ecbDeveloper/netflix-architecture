package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/ecbDeveloper/netflix-architecture/apps/content_processing_ms/internal/database"
	"github.com/ecbDeveloper/netflix-architecture/apps/content_processing_ms/internal/storage"
	"github.com/ecbDeveloper/netflix-architecture/apps/content_processing_ms/internal/transcoder"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

type ContentQueueType string

const (
	ContentQueueTypeMovie   ContentQueueType = "MOVIE"
	ContentQueueTypeEpisode ContentQueueType = "EPISODE"
)

type ContentProcessingMessage struct {
	ContentID   uuid.UUID        `json:"content_id"`
	ContentType ContentQueueType `json:"content_type"`
}

type Worker struct {
	repo       *database.Repository
	s3Client   *storage.S3Client
	transcoder *transcoder.Transcoder
	rabbitMQCh *amqp091.Channel
	queueName  string
	logger     *slog.Logger
}

func NewWorker(repo *database.Repository, s3Client *storage.S3Client, transcoder *transcoder.Transcoder, rabbitMQCh *amqp091.Channel, queueName string, logger *slog.Logger) *Worker {
	return &Worker{
		repo:       repo,
		s3Client:   s3Client,
		transcoder: transcoder,
		rabbitMQCh: rabbitMQCh,
		queueName:  queueName,
		logger:     logger,
	}
}

func (w *Worker) Start(ctx context.Context) {
	w.logger.Info("Content Processing Worker started. Consuming from RabbitMQ...")

	msgs, err := w.rabbitMQCh.Consume(
		w.queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		w.logger.Error("Failed to register a consumer", "error", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Worker shutting down...")
			return
		case d, ok := <-msgs:
			if !ok {
				w.logger.Info("RabbitMQ channel closed, stopping consumer...")
				return
			}
			w.processMessage(ctx, d)
		}
	}
}

func (w *Worker) processMessage(ctx context.Context, d amqp091.Delivery) {
	var payload ContentProcessingMessage
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		w.logger.Error("Failed to parse message", "error", err)
		d.Nack(false, false)
		return
	}

	var dbType database.ContentType
	switch payload.ContentType {
	case ContentQueueTypeMovie:
		dbType = database.TypeMovie
	case ContentQueueTypeEpisode:
		dbType = database.TypeEpisode
	default:
		w.logger.Error("Unknown content type", "type", payload.ContentType)
		d.Nack(false, false)
		return
	}

	status, err := w.repo.GetContentStatus(ctx, payload.ContentID, dbType)
	if err != nil {
		w.logger.Error("Failed to get content status", "error", err, "ID", payload.ContentID)
		d.Nack(false, true)
		return
	}

	if status == "PROCESSED" {
		w.logger.Info("Content already processed, skipping", "ID", payload.ContentID)
		d.Ack(false)
		return
	}

	w.logger.Info("Found pending content from queue", "ID", payload.ContentID, "Type", dbType)

	content := &database.PendingContent{
		ID:   payload.ContentID,
		Type: dbType,
	}

	err = w.processWithRetry(ctx, content)
	if err != nil {
		w.logger.Error("failed to process content", "ID", payload.ContentID, "error", err)
		d.Nack(false, false)
		return
	}

	w.logger.Info("Content successfully processed!", "ID", payload.ContentID)
	d.Ack(false)
}

func (w *Worker) processWithRetry(ctx context.Context, content *database.PendingContent) error {
	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		w.logger.Info("Processing content", "ID", content.ID, "Attempt", attempt, "MaxRetries", maxRetries)
		err := w.doProcess(ctx, content)
		if err == nil {
			return nil
		}

		lastErr = err
		w.logger.Error("Attempt failed", "Attempt", attempt, "error", err)

		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt*5) * time.Second)
		}
	}

	return fmt.Errorf("failed after %d attempts. Last error: %w", maxRetries, lastErr)
}

func (w *Worker) doProcess(ctx context.Context, content *database.PendingContent) error {
	idStr := content.ID.String()
	rawKey := fmt.Sprintf("raw/%s.mp4", idStr)
	processedPrefix := fmt.Sprintf("processed/%s", idStr)

	tempDir := filepath.Join(os.TempDir(), "netflix_processing", idStr)
	rawLocalPath := filepath.Join(tempDir, "raw.mp4")
	outLocalDir := filepath.Join(tempDir, "out")

	defer os.RemoveAll(tempDir)

	w.logger.Info("Downloading", "ID", idStr, "Key", rawKey)
	if err := w.s3Client.DownloadFile(ctx, rawKey, rawLocalPath); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	w.logger.Info("Transcoding to DASH", "ID", idStr)
	if err := w.transcoder.ProcessVideoToDASH(rawLocalPath, outLocalDir); err != nil {
		return fmt.Errorf("transcode failed: %w", err)
	}

	w.logger.Info("Uploading", "ID", idStr, "Prefix", processedPrefix)
	if err := w.s3Client.UploadDirectory(ctx, outLocalDir, processedPrefix); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	masterURL := fmt.Sprintf("%s/master.mpd", processedPrefix)
	w.logger.Info("Updating database with status PROCESSED", "ID", idStr)
	if err := w.repo.UpdateContentStatusAndURL(ctx, content.ID, content.Type, "PROCESSED", &masterURL); err != nil {
		return fmt.Errorf("database update failed: %w", err)
	}

	return nil
}
