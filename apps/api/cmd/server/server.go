package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/auth"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/episode"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/graph"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/graph/model"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/graph/resolvers"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/movie"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/profile"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/review"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/series"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/shared"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/user"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/watchhistory"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func init() {
	gob.Register(uuid.UUID{})
}

var userRoleOnDB = map[model.UserRole]int32{
	model.UserRoleAdmin:  shared.DBRoleAdmin,
	model.UserRoleMember: shared.DBRoleMember,
}

func main() {
	ctx := context.Background()

	loggerHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(loggerHandler)

	err := godotenv.Load()
	if err != nil {
		logger.Error("failed to load .env file", slog.Any("error", err))
	}

	port := os.Getenv("API_PORT")
	if port == "" {
		logger.Error("API_PORT is not set")
		os.Exit(1)
	}

	pool, err := initializeDatabaseConnection(ctx)
	if err != nil {
		logger.Error("failed to initialize db pool", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()

	redisPort := os.Getenv("REDIS_PORT")
	redisPass := os.Getenv("REDIS_PASS")
	redisHost := os.Getenv("REDIS_HOST")

	redisPool := &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		DialContext: func(ctx context.Context) (redis.Conn, error) {
			return redis.Dial("tcp", redisHost+":"+redisPort,
				redis.DialPassword(redisPass),
			)
		},
	}

	resolver, s, queries := initializeDependencies(pool, redisPool, logger)

	router := chi.NewRouter()
	router.Use(s.LoadAndSave)

	graphConfig := initializeGraphQLConfig(resolver, s, queries)

	srv := handler.New(graph.NewExecutableSchema(graphConfig))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	if os.Getenv("ENV") == "development" {
		srv.Use(extension.Introspection{})
	}
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	if os.Getenv("ENV") == "development" {
		router.Handle("/", playground.Handler("GraphQL playground", "/query"))
	}

	router.Handle("/query", srv)

	logger.Info("server initialized successfully", slog.String("url", "http://localhost:"+port+"/query"))
	if err := http.ListenAndServe(":"+port, router); err != nil {
		logger.Error("failed to start application", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func initializeDatabaseConnection(ctx context.Context) (*pgxpool.Pool, error) {
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
		return nil, fmt.Errorf("failed to create new db pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return pool, nil
}

func initializeDependencies(pool *pgxpool.Pool, redisPool *redis.Pool, logger *slog.Logger) (*resolvers.Resolver, *scs.SessionManager, *sqlc.Queries) {
	queries := sqlc.New(pool)

	userService := user.NewService(queries)
	episodeService := episode.NewService(queries)
	movieService := movie.NewService(queries)
	profileService := profile.NewService(queries)
	reviewService := review.NewService(queries)
	seriesService := series.NewService(queries)
	watchHistoryService := watchhistory.NewService(queries)
	authService := auth.NewService(queries)

	s := scs.New()
	s.Store = redisstore.New(redisPool)
	s.Lifetime = 24 * time.Hour
	s.Cookie.HttpOnly = true
	s.Cookie.SameSite = http.SameSiteLaxMode
	s.Cookie.Secure = os.Getenv("ENV") != "development"

	resolver := resolvers.NewResolver(
		logger,
		s,
		userService,
		episodeService,
		movieService,
		profileService,
		reviewService,
		seriesService,
		watchHistoryService,
		authService,
	)

	return resolver, s, queries
}

func initializeGraphQLConfig(resolver *resolvers.Resolver, s *scs.SessionManager, queries *sqlc.Queries) graph.Config {
	graphConfig := graph.Config{Resolvers: resolver}

	graphConfig.Directives.Auth = func(ctx context.Context, obj any, next graphql.Resolver) (res any, err error) {
		if !s.Exists(ctx, shared.SessionUserIDKey) {
			return nil, &gqlerror.Error{
				Message: "must be logged in",
				Extensions: map[string]any{
					"code": "UNAUTHENTICATED",
				},
			}
		}
		return next(ctx)
	}

	graphConfig.Directives.HasRole = func(ctx context.Context, obj any, next graphql.Resolver, role model.UserRole) (res any, err error) {
		dbRole, ok := userRoleOnDB[role]
		if !ok {
			return nil, &gqlerror.Error{
				Message: "access denied",
				Extensions: map[string]any{
					"code": "FORBIDDEN",
				},
			}
		}

		storedRoleID, ok := s.Get(ctx, shared.SessionRoleIDKey).(int32)
		if !ok || storedRoleID != dbRole {
			return nil, &gqlerror.Error{
				Message: "you don't have permission to perform this action",
				Extensions: map[string]any{
					"code": "FORBIDDEN",
				},
			}
		}

		return next(ctx)
	}

	graphConfig.Directives.ProfileSelectionIsRequired = func(ctx context.Context, obj any, next graphql.Resolver) (res any, err error) {
		userID, ok := s.Get(ctx, shared.SessionUserIDKey).(uuid.UUID)
		if !ok {
			return nil, &gqlerror.Error{
				Message: "authentication required",
				Extensions: map[string]any{
					"code": "UNAUTHENTICATED",
				},
			}
		}

		profileID, ok := s.Get(ctx, shared.SessionProfileIDKey).(uuid.UUID)
		if !ok {
			return nil, &gqlerror.Error{
				Message: "you must select a profile to access this content",
				Extensions: map[string]any{
					"code": "PROFILE_SELECTION_REQUIRED",
				},
			}
		}

		profileIsFromUser := false
		userProfiles, err := resolver.ProfileService.ListProfilesByUser(ctx, userID)
		if err != nil {
			return nil, &gqlerror.Error{
				Message: "internal error",
				Extensions: map[string]any{
					"code": "INTERNAL_SERVER",
				},
			}
		}

		for _, userProfile := range userProfiles {
			if userProfile.ID == profileID {
				profileIsFromUser = true
				break
			}
		}

		if err != nil || !profileIsFromUser {
			return nil, &gqlerror.Error{
				Message: "invalid profile selection",
				Extensions: map[string]any{
					"code": "FORBIDDEN",
				},
			}
		}

		return next(ctx)
	}

	return graphConfig
}
