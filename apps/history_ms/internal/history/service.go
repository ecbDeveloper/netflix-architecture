package history

import (
	"context"
	"errors"
	"fmt"

	"github.com/ecbDeveloper/netflix-architecture/apps/history_ms/internal/database/sqlc"
	pb "github.com/ecbDeveloper/netflix-architecture/proto/history"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedHistoryServiceServer
	queries *sqlc.Queries
}

func NewServer(queries *sqlc.Queries) pb.HistoryServiceServer {
	return &Server{
		queries: queries,
	}
}

func (s *Server) RecordWatch(ctx context.Context, req *pb.RecordWatchRequest) (*pb.WatchHistoryResponse, error) {
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

	params := sqlc.CreateWatchHistoryParams{
		ID:        whID,
		ProfileID: profileID,
	}

	if req.MovieId != nil {
		movieID, err := uuid.Parse(*req.MovieId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid movie_id")
		}
		params.MovieID = pgtype.UUID{Bytes: movieID, Valid: true}
	}

	if req.EpisodeId != nil {
		episodeID, err := uuid.Parse(*req.EpisodeId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid episode_id")
		}
		params.EpisodeID = pgtype.UUID{Bytes: episodeID, Valid: true}
	}

	if req.LastPositionSeconds != nil {
		params.LastPositionSeconds = pgtype.Int4{Int32: *req.LastPositionSeconds, Valid: true}
	}

	if req.IsCompleted != nil {
		params.IsCompleted = pgtype.Bool{Bool: *req.IsCompleted, Valid: true}
	}

	if req.GenreId != nil {
		params.GenreID = pgtype.Int4{Int32: *req.GenreId, Valid: true}
	}

	wh, err := s.queries.CreateWatchHistory(ctx, params)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to insert watch history: %v", err)
	}

	return toProto(wh), nil
}

func (s *Server) GetWatchHistory(ctx context.Context, req *pb.GetWatchHistoryRequest) (*pb.WatchHistoryResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	wh, err := s.queries.GetWatchHistory(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "watch history not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to fetch watch history: %v", err)
	}

	return toProto(wh), nil
}

func (s *Server) ListWatchHistory(ctx context.Context, req *pb.ListWatchHistoryRequest) (*pb.ListWatchHistoryResponse, error) {
	profileID, err := uuid.Parse(req.ProfileId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid profile_id")
	}

	histories, err := s.queries.ListWatchHistoryByProfile(ctx, profileID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list watch histories: %v", err)
	}

	result := make([]*pb.WatchHistoryResponse, len(histories))
	for i, wh := range histories {
		result[i] = toProto(wh)
	}

	return &pb.ListWatchHistoryResponse{Histories: result}, nil
}

func (s *Server) UpdateWatchProgress(ctx context.Context, req *pb.UpdateWatchProgressRequest) (*pb.WatchHistoryResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	current, err := s.queries.GetWatchHistory(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "watch history not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get watch history: %v", err)
	}

	params := sqlc.UpdateWatchProgressParams{
		ID:                  id,
		LastPositionSeconds: current.LastPositionSeconds,
		IsCompleted:         current.IsCompleted,
	}

	if req.LastPositionSeconds != nil {
		params.LastPositionSeconds = pgtype.Int4{Int32: *req.LastPositionSeconds, Valid: true}
	}
	if req.IsCompleted != nil {
		params.IsCompleted = pgtype.Bool{Bool: *req.IsCompleted, Valid: true}
	}

	wh, err := s.queries.UpdateWatchProgress(ctx, params)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update watch history: %v", err)
	}

	return toProto(wh), nil
}

func (s *Server) DeleteWatchHistory(ctx context.Context, req *pb.DeleteWatchHistoryRequest) (*pb.DeleteWatchHistoryResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	_, err = s.queries.GetWatchHistory(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "watch history not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to fetch watch history: %v", err)
	}

	if err := s.queries.DeleteWatchHistory(ctx, id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete watch history: %v", err)
	}

	return &pb.DeleteWatchHistoryResponse{Success: true}, nil
}

func (s *Server) GetMostWatched(ctx context.Context, req *pb.GetMostWatchedRequest) (*pb.MostWatchedResponse, error) {
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

	var items []*pb.MostWatchedItem

	for _, m := range movies {
		if m.MovieID.Valid {
			movieUUID, _ := uuid.FromBytes(m.MovieID.Bytes[:])
			items = append(items, &pb.MostWatchedItem{
				ContentId:   movieUUID.String(),
				ContentType: pb.ContentType_MOVIE,
				GenreId:     m.GenreID.Int32,
				WatchCount:  m.WatchCount,
			})
		}
	}

	for _, e := range episodes {
		if e.EpisodeID.Valid {
			episodeUUID, _ := uuid.FromBytes(e.EpisodeID.Bytes[:])
			items = append(items, &pb.MostWatchedItem{
				ContentId:   episodeUUID.String(),
				ContentType: pb.ContentType_SERIES,
				GenreId:     e.GenreID.Int32,
				WatchCount:  e.WatchCount,
			})
		}
	}

	return &pb.MostWatchedResponse{Items: items}, nil
}

func (s *Server) GetRecentlyWatched(ctx context.Context, req *pb.GetRecentlyWatchedRequest) (*pb.ListWatchHistoryResponse, error) {
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

	result := make([]*pb.WatchHistoryResponse, len(histories))
	for i, wh := range histories {
		result[i] = toProto(wh)
	}

	return &pb.ListWatchHistoryResponse{Histories: result}, nil
}

func toProto(wh sqlc.WatchHistory) *pb.WatchHistoryResponse {
	resp := &pb.WatchHistoryResponse{
		Id:                  wh.ID.String(),
		ProfileId:           wh.ProfileID.String(),
		LastPositionSeconds: wh.LastPositionSeconds.Int32,
		IsCompleted:         wh.IsCompleted.Bool,
		GenreId:             wh.GenreID.Int32,
	}

	if wh.MovieID.Valid {
		movieStr := fmt.Sprintf("%x-%x-%x-%x-%x",
			wh.MovieID.Bytes[0:4], wh.MovieID.Bytes[4:6], wh.MovieID.Bytes[6:8],
			wh.MovieID.Bytes[8:10], wh.MovieID.Bytes[10:16])
		resp.MovieId = &movieStr
	}

	if wh.EpisodeID.Valid {
		episodeStr := fmt.Sprintf("%x-%x-%x-%x-%x",
			wh.EpisodeID.Bytes[0:4], wh.EpisodeID.Bytes[4:6], wh.EpisodeID.Bytes[6:8],
			wh.EpisodeID.Bytes[8:10], wh.EpisodeID.Bytes[10:16])
		resp.EpisodeId = &episodeStr
	}

	if wh.WatchedAt.Valid {
		resp.WatchedAt = wh.WatchedAt.Time.String()
	}

	return resp
}
