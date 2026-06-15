package content

import (
	"time"

	"github.com/google/uuid"
)

type Content struct {
	ID             uuid.UUID
	Title          string
	Description    string
	ReleaseDate    time.Time
	MaturityRating MaturityRating
	ContentType    ContentType
	GenreID        int32
}

func (c Content) IsForKids() bool { return c.MaturityRating.IsKidsFriendly() }

func (c Content) IsAccessibleBy(hasParentalControls bool) bool {
	if hasParentalControls {
		return c.IsForKids()
	}
	return true
}

func (c Content) IsMovie() bool { return c.ContentType.IsMovie() }

func (c Content) IsSeries() bool { return c.ContentType.IsSeries() }
