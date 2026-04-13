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
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/watchhistory"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	Logger   *slog.Logger
	Sessions *scs.SessionManager

	UserService         user.Service
	EpisodeService      episode.Service
	MovieService        movie.Service
	ProfileService      profile.Service
	ReviewService       review.Service
	SeriesService       series.Service
	WatchHistoryService watchhistory.Service
	AuthService         auth.Service
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
	whs watchhistory.Service,
	as auth.Service,
) *Resolver {
	return &Resolver{
		Logger:              l,
		Sessions:            s,
		UserService:         us,
		EpisodeService:      es,
		MovieService:        ms,
		ProfileService:      ps,
		ReviewService:       rs,
		SeriesService:       ss,
		WatchHistoryService: whs,
		AuthService:         as,
	}
}
