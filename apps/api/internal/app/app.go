package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/auth"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/config"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/content"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/episode"
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
	"github.com/jackc/pgx/v5/pgxpool"
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

	s3Client, err := initializeS3Client(cfg)
	if err != nil {
		logger.Error("failed to initialize s3 client", slog.Any("error", err))
		os.Exit(1)
	}

	resolver, session := initializeDependencies(
		cfg,
		db,
		redisPool,
		logger,
		historyClient,
		recClient,
		s3Client,
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

func initializeDependencies(
	cfg *config.Config,
	pool *pgxpool.Pool,
	redisPool *redis.Pool,
	l *slog.Logger,
	historyClient historyv1.HistoryServiceClient,
	recommendationClient recommendationv1.RecommendationServiceClient,
	s3Client *s3.Client,
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
