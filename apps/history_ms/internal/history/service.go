package history

import (
	"context"
	"errors"

	"github.com/ecbDeveloper/netflix-architecture/apps/history_ms/internal/database/sqlc"
	commonpb "github.com/ecbDeveloper/netflix-architecture/gen/go/common/v1"
	historypb "github.com/ecbDeveloper/netflix-architecture/gen/go/history/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	historypb.UnimplementedHistoryServiceServer
	queries *sqlc.Queries
}

func NewServer(queries *sqlc.Queries) historypb.HistoryServiceServer {
	return &Server{
		queries: queries,
	}
}

func (s *Server) RecordWatch(ctx context.Context, req *historypb.RecordWatchHistoryRequest) (*historypb.RecordWatchHistoryResponse, error) {
	if req.MovieId == nil && req.EpisodeId == nil {
		return nil, status.Error(codes.InvalidArgument, "movie_id or episode_id is required")
	}
	if req.MovieId != nil && req.EpisodeId != nil {
		return nil, status.Error(codes.InvalidArgument, "provide only movie_id or episode_id, not both")
	}

	profileID, err := uuid.Parse(req.ProfileId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid profile_id")
	}

	whID := uuid.New()

	var lastPosition pgtype.Int4
	if req.LastPositionSeconds != nil {
		lastPosition = pgtype.Int4{Int32: *req.LastPositionSeconds, Valid: true}
	}

	var isCompleted pgtype.Bool
	if req.IsCompleted != nil {
		isCompleted = pgtype.Bool{Bool: *req.IsCompleted, Valid: true}
	}

	var wh sqlc.WatchHistory

	if req.MovieId != nil {
		movieID, err := uuid.Parse(*req.MovieId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid movie_id")
		}

		params := sqlc.UpsertMovieWatchHistoryParams{
			ID:                  whID,
			ProfileID:           profileID,
			MovieID:             pgtype.UUID{Bytes: movieID, Valid: true},
			GenreID:             pgtype.Int4{Int32: req.GenreId, Valid: true},
			LastPositionSeconds: lastPosition,
			IsCompleted:         isCompleted,
		}

		wh, err = s.queries.UpsertMovieWatchHistory(ctx, params)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to upsert movie watch history: %v", err)
		}
	}

	if req.EpisodeId != nil {
		episodeID, err := uuid.Parse(*req.EpisodeId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid episode_id")
		}

		params := sqlc.UpsertEpisodeWatchHistoryParams{
			ID:                  whID,
			ProfileID:           profileID,
			EpisodeID:           pgtype.UUID{Bytes: episodeID, Valid: true},
			GenreID:             pgtype.Int4{Int32: req.GenreId, Valid: true},
			LastPositionSeconds: lastPosition,
			IsCompleted:         isCompleted,
		}

		wh, err = s.queries.UpsertEpisodeWatchHistory(ctx, params)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to upsert episode watch history: %v", err)
		}
	}

	protoWH := toProto(wh)

	return &historypb.RecordWatchHistoryResponse{WatchHistory: protoWH}, nil
}

func (s *Server) GetWatchHistory(ctx context.Context, req *historypb.GetWatchHistoryRequest) (*historypb.GetWatchHistoryResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	profileID, err := uuid.Parse(req.ProfileId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid profile_id")
	}

	wh, err := s.queries.GetWatchHistory(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "watch history not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to fetch watch history: %v", err)
	}

	if wh.ProfileID != profileID {
		return nil, status.Error(codes.PermissionDenied, "you can't see watch history from others profiles")
	}

	protoWH := toProto(wh)

	return &historypb.GetWatchHistoryResponse{WatchHistory: protoWH}, nil
}

func (s *Server) ListWatchHistory(ctx context.Context, req *historypb.ListWatchHistoryRequest) (*historypb.ListWatchHistoryResponse, error) {
	profileID, err := uuid.Parse(req.ProfileId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid profile_id")
	}

	histories, err := s.queries.ListWatchHistoryByProfile(ctx, profileID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list watch histories: %v", err)
	}

	result := make([]*historypb.WatchHistory, len(histories))
	for i, wh := range histories {
		result[i] = toProto(wh)
	}

	return &historypb.ListWatchHistoryResponse{Histories: result}, nil
}

func (s *Server) DeleteWatchHistory(ctx context.Context, req *historypb.DeleteWatchHistoryRequest) (*historypb.DeleteWatchHistoryResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	profileID, err := uuid.Parse(req.ProfileId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid profile_id")
	}

	wh, err := s.queries.GetWatchHistory(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "watch history not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to fetch watch history: %v", err)
	}

	if wh.ProfileID != profileID {
		return nil, status.Error(codes.PermissionDenied, "you can't see watch history from others profiles")
	}

	if err := s.queries.DeleteWatchHistory(ctx, id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete watch history: %v", err)
	}

	return &historypb.DeleteWatchHistoryResponse{Success: true}, nil
}

func (s *Server) GetMostWatched(ctx context.Context, req *historypb.GetMostWatchedRequest) (*historypb.GetMostWatchedResponse, error) {
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}

	movies, err := s.queries.GetMostWatchedMovies(ctx, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get most watched movies: %v", err)
	}

	episodes, err := s.queries.GetMostWatchedEpisodes(ctx, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get most watched episodes: %v", err)
	}

	var items []*historypb.MostWatchedItem

	for _, m := range movies {
		if m.MovieID.Valid {
			movieUUID, _ := uuid.FromBytes(m.MovieID.Bytes[:])
			items = append(items, &historypb.MostWatchedItem{
				ContentId:   movieUUID.String(),
				ContentType: commonpb.ContentType_CONTENT_TYPE_MOVIE,
				GenreId:     m.GenreID.Int32,
				WatchCount:  m.WatchCount,
			})
		}
	}

	for _, e := range episodes {
		if e.EpisodeID.Valid {
			episodeUUID, _ := uuid.FromBytes(e.EpisodeID.Bytes[:])
			items = append(items, &historypb.MostWatchedItem{
				ContentId:   episodeUUID.String(),
				ContentType: commonpb.ContentType_CONTENT_TYPE_SERIES,
				GenreId:     e.GenreID.Int32,
				WatchCount:  e.WatchCount,
			})
		}
	}

	return &historypb.GetMostWatchedResponse{Items: items}, nil
}

func (s *Server) GetRecentlyWatched(ctx context.Context, req *historypb.GetRecentlyWatchedRequest) (*historypb.GetRecentlyWatchedResponse, error) {
	profileID, err := uuid.Parse(req.ProfileId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid profile_id")
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}

	histories, err := s.queries.GetRecentlyWatchedByProfile(ctx, sqlc.GetRecentlyWatchedByProfileParams{
		ProfileID: profileID,
		Limit:     limit,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get recently watched: %v", err)
	}

	result := make([]*historypb.WatchHistory, len(histories))
	for i, wh := range histories {
		result[i] = toProto(wh)
	}

	return &historypb.GetRecentlyWatchedResponse{Histories: result}, nil
}

func toProto(wh sqlc.WatchHistory) *historypb.WatchHistory {
	resp := &historypb.WatchHistory{
		Id:                  wh.ID.String(),
		ProfileId:           wh.ProfileID.String(),
		LastPositionSeconds: wh.LastPositionSeconds.Int32,
		IsCompleted:         wh.IsCompleted.Bool,
		GenreId:             wh.GenreID.Int32,
	}

	if wh.MovieID.Valid {
		movieIDStr := wh.MovieID.String()

		resp.MovieId = &movieIDStr
	}

	if wh.EpisodeID.Valid {
		episodeIDStr := wh.EpisodeID.String()
		resp.EpisodeId = &episodeIDStr
	}

	if wh.WatchedAt.Valid {
		resp.WatchedAt = wh.WatchedAt.Time.String()
	}

	return resp
}
