package resolvers

import (
	"log/slog"

	"github.com/alexedwards/scs/v2"
	"github.com/ecbDeveloper/netflix-architecture/internal/auth"
	"github.com/ecbDeveloper/netflix-architecture/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/internal/episode"
	"github.com/ecbDeveloper/netflix-architecture/internal/movie"
	"github.com/ecbDeveloper/netflix-architecture/internal/profile"
	"github.com/ecbDeveloper/netflix-architecture/internal/review"
	"github.com/ecbDeveloper/netflix-architecture/internal/series"
	"github.com/ecbDeveloper/netflix-architecture/internal/user"
	"github.com/ecbDeveloper/netflix-architecture/internal/watchhistory"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	Queries  *sqlc.Queries
	Logger   *slog.Logger
	Sessions *scs.SessionManager

	UserService         *user.Service
	EpisodeService      *episode.Service
	MovieService        *movie.Service
	ProfileService      *profile.Service
	ReviewService       *review.Service
	SeriesService       *series.Service
	WatchhistoryService *watchhistory.Service
	AuthService         *auth.Service
}

func NewResolver(
	q *sqlc.Queries,
	l *slog.Logger,
	s *scs.SessionManager,
	us *user.Service,
	es *episode.Service,
	ms *movie.Service,
	ps *profile.Service,
	rs *review.Service,
	ss *series.Service,
	whs *watchhistory.Service,
	as *auth.Service,
) *Resolver {
	return &Resolver{
		Logger:              l,
		Sessions:            s,
		Queries:             q,
		UserService:         us,
		EpisodeService:      es,
		MovieService:        ms,
		ProfileService:      ps,
		ReviewService:       rs,
		SeriesService:       ss,
		WatchhistoryService: whs,
		AuthService:         as,
	}
}
