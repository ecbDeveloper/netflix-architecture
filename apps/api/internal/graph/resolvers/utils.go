package resolvers

import (
	"context"

	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/apperror"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/graph/model"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/shared"
	historypb "github.com/ecbDeveloper/netflix-architecture/proto/history"
	recommendationpb "github.com/ecbDeveloper/netflix-architecture/proto/recommendation"
	"github.com/google/uuid"
)

var hContentTypeToGraph = map[historypb.ContentType]model.ContentType{
	1: model.ContentTypeMovie,
	2: model.ContentTypeSeries,
}
var recContentTypeToGraph = map[recommendationpb.ContentType]model.ContentType{
	1: model.ContentTypeMovie,
	2: model.ContentTypeSeries,
}

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

func protoToWatchHistory(resp *historypb.WatchHistoryResponse) *model.WatchHistory {
	wh := &model.WatchHistory{
		WatchedAt: resp.WatchedAt,
	}

	whID, _ := uuid.Parse(resp.Id)
	wh.ID = whID

	profileID, _ := uuid.Parse(resp.ProfileId)
	wh.ProfileID = profileID

	if resp.MovieId != nil {
		movieID, _ := uuid.Parse(*resp.MovieId)
		wh.MovieID = &movieID
	}
	if resp.EpisodeId != nil {
		episodeID, _ := uuid.Parse(*resp.EpisodeId)
		wh.EpisodeID = &episodeID
	}

	lps := resp.LastPositionSeconds
	wh.LastPositionSeconds = &lps

	ic := resp.IsCompleted
	wh.IsCompleted = &ic

	return wh
}
