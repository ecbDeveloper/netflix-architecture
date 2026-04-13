package resolvers

import (
	"context"

	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/apperror"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/shared"
	"github.com/google/uuid"
)

func (r *Resolver) getUserIDFromSession(ctx context.Context) (uuid.UUID, error) {
	userID, ok := r.Sessions.Get(ctx, shared.SessionUserIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, &apperror.UnauthorizedError{Message: ErrUserNotLoggedIn.Error()}
	}
	return userID, nil
}

func (r *Resolver) getUserRoleIDFromSession(ctx context.Context) (int32, error) {
	roleID, ok := r.Sessions.Get(ctx, shared.SessionRoleIDKey).(int32)
	if !ok {
		return 0, &apperror.UnauthorizedError{Message: ErrUserNotLoggedIn.Error()}
	}
	return roleID, nil
}

func (r *Resolver) getProfileIDFromSession(ctx context.Context) (uuid.UUID, error) {
	profileID, ok := r.Sessions.Get(ctx, shared.SessionProfileIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, &apperror.ForbiddenError{Message: "you must select a profile to access this content"}
	}
	return profileID, nil
}
