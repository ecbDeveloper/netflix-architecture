package profile

import "github.com/google/uuid"

type Profile struct {
	ID                  uuid.UUID
	UserID              uuid.UUID
	Name                string
	HasParentalControls bool
}

func (p Profile) BelongsTo(userID uuid.UUID) bool {
	return p.UserID == userID
}

func (p Profile) HasRestrictedAccess() bool {
	return p.HasParentalControls
}
