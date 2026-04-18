package resolvers

import (
	"log/slog"

	"github.com/alexedwards/scs/v2"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/auth"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/content"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/episode"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/profile"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/review"
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
	ProfileService       profile.Service
	ReviewService        review.Service
	AuthService          auth.Service
	ContentService       content.Service
	HistoryClient        historypb.HistoryServiceClient
	RecommendationClient recommendationpb.RecommendationServiceClient
}

func NewResolver(
	l *slog.Logger,
	s *scs.SessionManager,
	us user.Service,
	es episode.Service,
	ps profile.Service,
	rs review.Service,
	as auth.Service,
	cs content.Service,
	hc historypb.HistoryServiceClient,
	rc recommendationpb.RecommendationServiceClient,
) *Resolver {
	return &Resolver{
		Logger:               l,
		Sessions:             s,
		UserService:          us,
		EpisodeService:       es,
		ProfileService:       ps,
		ReviewService:        rs,
		AuthService:          as,
		ContentService:       cs,
		HistoryClient:        hc,
		RecommendationClient: rc,
	}
}
