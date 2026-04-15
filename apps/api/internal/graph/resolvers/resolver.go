package resolvers

import (
	"log/slog"

	"github.com/alexedwards/scs/v2"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/auth"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/episode"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/movie"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/profile"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/review"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/series"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/user"
	historypb "github.com/ecbDeveloper/netflix-architecture/proto/history"
	recommendationpb "github.com/ecbDeveloper/netflix-architecture/proto/recommendation"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	Logger   *slog.Logger
	Sessions *scs.SessionManager

	UserService          user.Service
	EpisodeService       episode.Service
	MovieService         movie.Service
	ProfileService       profile.Service
	ReviewService        review.Service
	SeriesService        series.Service
	AuthService          auth.Service
	HistoryClient        historypb.HistoryServiceClient
	RecommendationClient recommendationpb.RecommendationServiceClient
}

func NewResolver(
	l *slog.Logger,
	s *scs.SessionManager,
	us user.Service,
	es episode.Service,
	ms movie.Service,
	ps profile.Service,
	rs review.Service,
	ss series.Service,
	as auth.Service,
	hc historypb.HistoryServiceClient,
	rc recommendationpb.RecommendationServiceClient,
) *Resolver {
	return &Resolver{
		Logger:               l,
		Sessions:             s,
		UserService:          us,
		EpisodeService:       es,
		MovieService:         ms,
		ProfileService:       ps,
		ReviewService:        rs,
		SeriesService:        ss,
		AuthService:          as,
		HistoryClient:        hc,
		RecommendationClient: rc,
	}
}
