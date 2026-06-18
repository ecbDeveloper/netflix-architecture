package server

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
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/graph/resolvers"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/infra"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/infra/storage"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/middleware"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/profile"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/review"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/user"
	historyv1 "github.com/ecbDeveloper/netflix-architecture/gen/go/history/v1"
	recommendationv1 "github.com/ecbDeveloper/netflix-architecture/gen/go/recommendation/v1"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rabbitmq/amqp091-go"
)

func Run(ctx context.Context, logger *slog.Logger, cfg *config.Config) {
	db, err := infra.InitializeDatabase(ctx, cfg)
	if err != nil {
		logger.Error("failed to initialize database", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	redisPool, err := infra.InitializeRedis(cfg)
	if err != nil {
		logger.Error("failed to initialize redis", slog.Any("error", err))
		os.Exit(1)
	}
	defer redisPool.Close()

	s3Client, err := infra.InitializeS3Client(cfg)
	if err != nil {
		logger.Error("failed to initialize s3 client", slog.Any("error", err))
		os.Exit(1)
	}

	historyClient, recClient, err := infra.InitializeGRPC(cfg)
	if err != nil {
		logger.Error("failed to initialize grpc clients", slog.Any("error", err))
		os.Exit(1)
	}

	rabbitMQConn, err := infra.InitializeRabbitMQ(cfg)
	if err != nil {
		logger.Error("failed to initialize rabbitMQ", slog.Any("error", err))
		os.Exit(1)
	}
	defer rabbitMQConn.Close()

	rabbitMQCh, err := rabbitMQConn.Channel()
	if err != nil {
		logger.Error("failed to get rabbitmq channel: %w", err)
		os.Exit(1)
	}
	defer rabbitMQCh.Close()

	if err := infra.DeclareContentQueue(cfg, rabbitMQCh); err != nil {
		logger.Error("failed to declare content queue: %w", err)
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
		rabbitMQCh,
	)
	resolver.Logger = logger

	graphConfig := initializeGraphQLConfig(resolver, session)
	graphServer := buildGraphQLServer(graphConfig, cfg)

	router := chi.NewRouter()
	router.Use(session.LoadAndSave, middleware.RequestLogger(logger))

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
	rabbitMQCh *amqp091.Channel,
) (*resolvers.Resolver, *scs.SessionManager) {
	queries := sqlc.New(pool)

	storageService := storage.NewService(s3Client, cfg.S3BucketName, cfg.S3EndPointURL)
	userService := user.NewService(queries)
	profileService := profile.NewService(queries)
	authService := auth.NewService(queries, userService)
	episodeService := episode.NewService(queries, storageService, profileService, rabbitMQCh)
	contentService := content.NewService(queries, pool, storageService, profileService, rabbitMQCh)
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
