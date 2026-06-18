package shared

import "github.com/google/uuid"

type ContentQueueType string

const (
	ContentQueueTypeMovie   ContentQueueType = "MOVIE"
	ContentQueueTypeEpisode ContentQueueType = "EPISODE"
)

type ContentProcessingMessage struct {
	ContentID   uuid.UUID        `json:"content_id"`
	ContentType ContentQueueType `json:"content_type"`
}
