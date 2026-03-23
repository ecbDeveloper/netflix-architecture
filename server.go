package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/ecbDeveloper/netflix-architecture/internal/auth"
	"github.com/ecbDeveloper/netflix-architecture/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/internal/episode"
	"github.com/ecbDeveloper/netflix-architecture/internal/graph"
	"github.com/ecbDeveloper/netflix-architecture/internal/graph/resolvers"
	"github.com/ecbDeveloper/netflix-architecture/internal/movie"
	"github.com/ecbDeveloper/netflix-architecture/internal/profile"
	"github.com/ecbDeveloper/netflix-architecture/internal/review"
	"github.com/ecbDeveloper/netflix-architecture/internal/series"
	"github.com/ecbDeveloper/netflix-architecture/internal/user"
	"github.com/ecbDeveloper/netflix-architecture/internal/watchhistory"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vektah/gqlparser/v2/ast"
)

func init() {
	gob.Register(uuid.UUID{})
}

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	router := chi.NewRouter()

	loggerHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(loggerHandler)

	ctx := context.Background()
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_NAME"),
	)

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		logger.Error("failed to initialize db pool", slog.String("error", err.Error()))
		os.Exit(1)
	}
	queries := sqlc.New(pool)

	s := scs.New()
	s.Store = pgxstore.New(pool)
	s.Lifetime = 24 * time.Hour
	s.Cookie.HttpOnly = true
	s.Cookie.SameSite = http.SameSiteLaxMode

	userService := user.NewService(queries)
	episodeService := episode.NewService(queries)
	movieService := movie.NewService(queries)
	profileService := profile.NewService(queries)
	reviewService := review.NewService(queries)
	seriesService := series.NewService(queries)
	watchhistoryService := watchhistory.NewService(queries)
	authService := auth.NewService(queries)

	router.Use(s.LoadAndSave)

	graphConfig := graph.Config{Resolvers: &resolvers.Resolver{
		Queries:             queries,
		Logger:              logger,
		Sessions:            s,
		UserService:         userService,
		EpisodeService:      episodeService,
		MovieService:        movieService,
		ProfileService:      profileService,
		ReviewService:       reviewService,
		SeriesService:       seriesService,
		WatchhistoryService: watchhistoryService,
		AuthService:         authService,
	}}

	srv := handler.New(graph.NewExecutableSchema(graphConfig))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	router.Handle("/", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", srv)

	logger.Info("server initialized successfully", slog.String("url", "http://localhost:"+port+"/query"))
	if err := http.ListenAndServe(":"+port, router); err != nil {
		logger.Error("failed to start application", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
