package recommendation

import (
	"context"
	"sort"

	"github.com/ecbDeveloper/netflix-architecture/apps/recommendation_ms/internal/database/sqlc"
	historypb "github.com/ecbDeveloper/netflix-architecture/proto/history"
	pb "github.com/ecbDeveloper/netflix-architecture/proto/recommendation"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedRecommendationServiceServer
	queries       *sqlc.Queries
	historyClient historypb.HistoryServiceClient
}

func NewServer(queries *sqlc.Queries, historyClient historypb.HistoryServiceClient) pb.RecommendationServiceServer {
	return &Server{
		queries:       queries,
		historyClient: historyClient,
	}
}

type scoredContent struct {
	contentID   string
	contentType pb.ContentType
	score       float64
	reason      string
}

func (s *Server) GetRecommendations(ctx context.Context, req *pb.GetRecommendationsRequest) (*pb.GetRecommendationsResponse, error) {
	profileID, err := uuid.Parse(req.ProfileId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid profile_id")
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20
	}

	historyResp, err := s.historyClient.ListWatchHistory(ctx, &historypb.ListWatchHistoryRequest{
		ProfileId: profileID.String(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch watch history: %v", err)
	}

	watchedSet := make(map[string]bool)
	watchedGenres := make(map[int32]int)
	for _, h := range historyResp.Histories {
		if h.MovieId != nil {
			watchedSet[*h.MovieId] = true
		}
		if h.EpisodeId != nil {
			watchedSet[*h.EpisodeId] = true
		}
		if h.GenreId > 0 {
			watchedGenres[h.GenreId]++
		}
	}

	mostWatchedResp, err := s.historyClient.GetMostWatched(ctx, &historypb.GetMostWatchedRequest{
		Limit: int32(limit * 3),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch most watched: %v", err)
	}

	scored := make([]scoredContent, 0)

	for _, item := range mostWatchedResp.Items {
		if watchedSet[item.ContentId] {
			continue
		}
		scored = append(scored, scoredContent{
			contentID:   item.ContentId,
			contentType: pb.ContentType(item.ContentType),
			score:       float64(item.WatchCount) * 1.0,
			reason:      "most_watched",
		})
	}

	for i := range scored {
		for _, item := range mostWatchedResp.Items {
			if item.ContentId == scored[i].contentID && item.GenreId > 0 {
				if count, ok := watchedGenres[item.GenreId]; ok {
					scored[i].score += float64(count) * 2.0
					scored[i].reason = "same_genre"
				}
			}
		}
	}

	for _, topRated := range req.TopRatedContents {
		if watchedSet[topRated.ContentId] {
			continue
		}

		found := false
		for i := range scored {
			if scored[i].contentID == topRated.ContentId {
				scored[i].score += topRated.AvgRating * 3.0
				if scored[i].reason != "same_genre" {
					scored[i].reason = "top_rated"
				}
				found = true
				break
			}
		}
		if !found {
			s := scoredContent{
				contentID:   topRated.ContentId,
				contentType: topRated.ContentType,
				score:       topRated.AvgRating * 3.0,
				reason:      "top_rated",
			}
			if count, ok := watchedGenres[topRated.GenreId]; ok {
				s.score += float64(count) * 2.0
				s.reason = "same_genre"
			}
			scored = append(scored, s)
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	if len(scored) > limit {
		scored = scored[:limit]
	}

	_ = s.queries.DeleteRecommendationsByProfile(ctx, profileID)
	for _, sc := range scored {
		contentUUID, err := uuid.Parse(sc.contentID)
		if err != nil {
			continue
		}
		var dbContentType sqlc.ContentTypeEnum
		if sc.contentType == pb.ContentType_MOVIE {
			dbContentType = sqlc.ContentTypeEnumMOVIE
		} else if sc.contentType == pb.ContentType_SERIES {
			dbContentType = sqlc.ContentTypeEnumSERIES
		}

		_, _ = s.queries.CreateRecommendation(ctx, sqlc.CreateRecommendationParams{
			ID:          uuid.New(),
			ProfileID:   profileID,
			ContentID:   contentUUID,
			ContentType: dbContentType,
			Score:       sc.score,
			Reason:      sc.reason,
		})
	}

	recommendations := make([]*pb.RecommendedContent, len(scored))
	for i, sc := range scored {
		recommendations[i] = &pb.RecommendedContent{
			ContentId:   sc.contentID,
			ContentType: sc.contentType,
			Score:       sc.score,
			Reason:      sc.reason,
		}
	}

	return &pb.GetRecommendationsResponse{
		Recommendations: recommendations,
	}, nil
}
