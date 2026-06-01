package app

import (
	"context"
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
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/config"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/content"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/episode"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/graph"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/graph/model"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/graph/resolvers"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/middleware"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/profile"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/review"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/shared"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/storage"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/user"
	historyv1 "github.com/ecbDeveloper/netflix-architecture/gen/go/history/v1"
	recommendationv1 "github.com/ecbDeveloper/netflix-architecture/gen/go/recommendation/v1"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var userRoleOnDB = map[model.UserRole]int32{
	model.UserRoleAdmin:  shared.DBRoleAdmin,
	model.UserRoleMember: shared.DBRoleMember,
}

func Run(ctx context.Context, logger *slog.Logger, cfg *config.Config) {
	db, err := initializeDatabase(ctx, cfg)
	if err != nil {
		logger.Error("failed to initialize database", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	redisPool, err := initializeRedis(cfg)
	if err != nil {
		logger.Error("failed to initialize redis", slog.Any("error", err))
		os.Exit(1)
	}
	defer redisPool.Close()

	historyAddr := os.Getenv("HISTORY_GRPC_ADDR")
	historyConn, err := grpc.NewClient(
		historyAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Error("failed to connect to history ms", slog.Any("error", err))
		os.Exit(1)
	}
	defer historyConn.Close()
	historyClient := historyv1.NewHistoryServiceClient(historyConn)

	recAddr := os.Getenv("RECOMMENDATION_GRPC_ADDR")
	recConn, err := grpc.NewClient(
		recAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Error("failed to connect to recommendation ms", slog.Any("error", err))
		os.Exit(1)
	}
	defer recConn.Close()
	recClient := recommendationv1.NewRecommendationServiceClient(recConn)

	resolver, session := initializeDependencies(
		cfg,
		db,
		redisPool,
		logger,
		historyClient,
		recClient,
	)
	resolver.Logger = logger

	graphConfig := initializeGraphQLConfig(resolver, session)
	graphServer := buildGraphQLServer(graphConfig, cfg)

	router := chi.NewRouter()
	router.Use(session.LoadAndSave, middleware.RequestLogger(logger))

	router.Handle("/storage/*", http.StripPrefix("/storage/", http.FileServer(http.Dir(cfg.UploadPath))))
	router.Handle("/query", graphServer)

	if cfg.IsDevelopment() {
		router.Handle("/", playground.Handler("GraphQL playground", "/query"))
	}

	logger.Info("server initialized successfully", slog.String("url", "http://localhost:"+cfg.APIPort+"/query"))

	if err := http.ListenAndServe(":"+cfg.APIPort, router); err != nil {
		logger.Error("failed to start server", slog.Any("error", err))
		os.Exit(1)
	}
}

func buildGraphQLServer(cfg graph.Config, appCfg *config.Config) *handler.Server {
	srv := handler.New(graph.NewExecutableSchema(cfg))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{
		MaxUploadSize: shared.MaxUploadSize,
		MaxMemory:     shared.MaxMemory,
	})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	srv.Use(extension.AutomaticPersistedQuery{Cache: lru.New[string](100)})

	if appCfg.IsDevelopment() {
		srv.Use(extension.Introspection{})
	}

	return srv
}

func initializeDependencies(
	cfg *config.Config,
	pool *pgxpool.Pool,
	redisPool *redis.Pool,
	l *slog.Logger,
	historyClient historyv1.HistoryServiceClient,
	recommendationClient recommendationv1.RecommendationServiceClient,
) (*resolvers.Resolver, *scs.SessionManager) {
	queries := sqlc.New(pool)

	storageService := storage.NewService(cfg.UploadPath)
	userService := user.NewService(queries)
	profileService := profile.NewService(queries)
	authService := auth.NewService(queries, userService)
	episodeService := episode.NewService(queries, storageService, profileService)
	contentService := content.NewService(queries, pool, storageService, profileService)
	reviewService := review.NewService(queries, episodeService)

	s := scs.New()
	s.Store = redisstore.New(redisPool)
	s.Lifetime = 24 * time.Hour
	s.Cookie.HttpOnly = true
	s.Cookie.SameSite = http.SameSiteLaxMode
	s.Cookie.Secure = !cfg.IsDevelopment()

	resolver := resolvers.NewResolver(
		l,
		s,
		userService,
		episodeService,
		profileService,
		reviewService,
		authService,
		contentService,
		historyClient,
		recommendationClient,
	)

	return resolver, s
}

func initializeGraphQLConfig(resolver *resolvers.Resolver, s *scs.SessionManager) graph.Config {
	graphConfig := graph.Config{Resolvers: resolver}

	graphConfig.Directives.Auth = func(ctx context.Context, obj any, next graphql.Resolver) (res any, err error) {
		if !s.Exists(ctx, shared.SessionUserIDKey) {
			return nil, &gqlerror.Error{
				Message:    "must be logged in",
				Extensions: map[string]any{"code": "UNAUTHENTICATED"},
			}
		}
		return next(ctx)
	}

	graphConfig.Directives.HasRole = func(ctx context.Context, obj any, next graphql.Resolver, role model.UserRole) (res any, err error) {
		dbRole, ok := userRoleOnDB[role]
		if !ok {
			return nil, &gqlerror.Error{
				Message:    "access denied",
				Extensions: map[string]any{"code": "FORBIDDEN"},
			}
		}

		storedRoleID, ok := s.Get(ctx, shared.SessionRoleIDKey).(int32)
		if !ok || storedRoleID != dbRole {
			return nil, &gqlerror.Error{
				Message:    "you don't have permission to perform this action",
				Extensions: map[string]any{"code": "FORBIDDEN"},
			}
		}

		return next(ctx)
	}

	graphConfig.Directives.ProfileSelectionIsRequired = func(ctx context.Context, obj any, next graphql.Resolver) (res any, err error) {
		userID, ok := s.Get(ctx, shared.SessionUserIDKey).(uuid.UUID)
		if !ok {
			return nil, &gqlerror.Error{
				Message:    "authentication required",
				Extensions: map[string]any{"code": "UNAUTHENTICATED"},
			}
		}

		profileID, ok := s.Get(ctx, shared.SessionProfileIDKey).(uuid.UUID)
		if !ok {
			return nil, &gqlerror.Error{
				Message:    "you must select a profile to access this content",
				Extensions: map[string]any{"code": "PROFILE_SELECTION_REQUIRED"},
			}
		}

		userProfiles, err := resolver.ProfileService.ListProfilesByUser(ctx, userID)
		if err != nil {
			return nil, &gqlerror.Error{
				Message:    "internal error",
				Extensions: map[string]any{"code": "INTERNAL_SERVER"},
			}
		}

		for _, p := range userProfiles {
			if p.ID == profileID {
				return next(ctx)
			}
		}

		return nil, &gqlerror.Error{
			Message:    "invalid profile selection",
			Extensions: map[string]any{"code": "FORBIDDEN"},
		}
	}

	return graphConfig
}

func initializeDatabase(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to create db pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return pool, nil
}

func initializeRedis(cfg *config.Config) (*redis.Pool, error) {
	pool := &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		DialContext: func(context.Context) (redis.Conn, error) {
			return redis.Dial("tcp", cfg.RedisAddr(), redis.DialPassword(cfg.RedisPass))
		},
	}

	conn := pool.Get()
	defer conn.Close()

	if err := conn.Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	if _, err := conn.Do("PING"); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return pool, nil
}
