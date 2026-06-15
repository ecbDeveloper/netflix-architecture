package review

import "github.com/google/uuid"

type Review struct {
	ID        uuid.UUID
	ProfileID uuid.UUID
	Rating    Rating
	Comment   *string
	MovieID   *uuid.UUID
	EpisodeID *uuid.UUID
}

func (r Review) BelongsTo(profileID uuid.UUID) bool { return r.ProfileID == profileID }

func (r Review) IsMovieReview() bool { return r.MovieID != nil }

func (r Review) IsEpisodeReview() bool { return r.EpisodeID != nil }
