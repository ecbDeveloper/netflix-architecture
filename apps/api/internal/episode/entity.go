package episode

import "github.com/google/uuid"

type Episode struct {
	ID              uuid.UUID
	SeriesID        uuid.UUID
	Season          Season
	EpisodeNumber   EpisodeNumber
	Title           string
	DurationSeconds Duration
	ContentURL      string
}

func (e Episode) BelongsToSeries(seriesID uuid.UUID) bool {
	return e.SeriesID == seriesID
}
