package resolvers

import (
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/graph/model"
	commonv1 "github.com/ecbDeveloper/netflix-architecture/gen/go/common/v1"
	historyv1 "github.com/ecbDeveloper/netflix-architecture/gen/go/history/v1"
	"github.com/google/uuid"
)

var protoContentTypeToGraphQL = map[commonv1.ContentType]model.ContentType{
	commonv1.ContentType_CONTENT_TYPE_MOVIE:  model.ContentTypeMovie,
	commonv1.ContentType_CONTENT_TYPE_SERIES: model.ContentTypeSeries,
}

func protoWatchHistoryToGraphQL(resp *historyv1.WatchHistory) *model.WatchHistory {
	wh := &model.WatchHistory{
		WatchedAt: resp.WatchedAt,
	}

	whID, _ := uuid.Parse(resp.Id)
	wh.ID = whID

	if resp.MovieId != nil {
		movieID, _ := uuid.Parse(*resp.MovieId)
		wh.MovieID = movieID
	}
	if resp.EpisodeId != nil {
		episodeID, _ := uuid.Parse(*resp.EpisodeId)
		wh.EpisodeID = episodeID
	}

	lps := resp.LastPositionSeconds
	wh.LastPositionSeconds = &lps

	ic := resp.IsCompleted
	wh.IsCompleted = &ic

	return wh
}
