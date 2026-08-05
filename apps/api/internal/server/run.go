package server

import (
	"context"
	"errors"
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
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/infra/queue"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/infra/storage"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/profile"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/review"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/shared"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/user"
	historyv1 "github.com/ecbDeveloper/netflix-architecture/gen/go/history/v1"
	recommendationv1 "github.com/ecbDeveloper/netflix-architecture/gen/go/recommendation/v1"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rabbitmq/amqp091-go"
	"github.com/riandyrn/otelchi"
	otelchimetric "github.com/riandyrn/otelchi/metric"
)

func Run(ctx context.Context, logger *slog.Logger, appConfig *config.Config) {
	db, err := infra.InitializeDatabase(ctx, appConfig)
	if err != nil {
		logger.Error("failed to initialize database", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	redisPool, err := infra.InitializeRedis(appConfig)
	if err != nil {
		logger.Error("failed to initialize redis", slog.Any("error", err))
		os.Exit(1)
	}
	defer redisPool.Close()

	s3Client, err := infra.InitializeS3Client(appConfig)
	if err != nil {
		logger.Error("failed to initialize s3 client", slog.Any("error", err))
		os.Exit(1)
	}

	historyClient, recClient, cleanup, err := infra.InitializeGRPC(appConfig)
	if err != nil {
		logger.Error("failed to initialize grpc clients", slog.Any("error", err))
		os.Exit(1)
	}
	defer cleanup()

	rabbitMQConn, err := infra.InitializeRabbitMQ(appConfig)
	if err != nil {
		logger.Error("failed to initialize rabbitMQ", slog.Any("error", err))
		os.Exit(1)
	}
	defer rabbitMQConn.Close()

	rabbitMQCh, err := rabbitMQConn.Channel()
	if err != nil {
		logger.Error("failed to get rabbitmq channel", slog.Any("error", err))
		os.Exit(1)
	}
	defer rabbitMQCh.Close()

	if err := infra.DeclareContentQueue(appConfig, rabbitMQCh); err != nil {
		logger.Error("failed to declare content queue", slog.Any("error", err))
		os.Exit(1)
	}

	meterProvider, otelShutdown, err := infra.SetupOTelSDK(ctx)
	if err != nil {
		logger.Error("failed to initialize open telemetry sdk", slog.Any("error", err))
		os.Exit(1)
	}
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	otelConfig := otelchimetric.NewBaseConfig(
		shared.ServerName,
		otelchimetric.WithMeterProvider(meterProvider),
	)

	resolver, session := initializeDependencies(
		appConfig,
		db,
		redisPool,
		logger,
		historyClient,
		recClient,
		s3Client,
		rabbitMQCh,
	)

	graphConfig := initializeGraphQLConfig(resolver, session)
	graphServer := buildGraphQLServer(graphConfig, appConfig, logger)

	router := chi.NewRouter()
	router.Use(
		session.LoadAndSave,
		chimiddleware.Recoverer,
		otelchi.Middleware(
			shared.ServerName,
			otelchi.WithChiRoutes(router),
		),
		otelchimetric.NewServerRequestDuration(otelConfig),
		otelchimetric.NewServerActiveRequests(otelConfig),
		otelchimetric.NewServerRequestBodySize(otelConfig),
		otelchimetric.NewServerResponseBodySize(otelConfig),
	)

	router.Handle("/query", graphServer)

	if appConfig.IsDevelopment() {
		router.Handle("/", playground.Handler("GraphQL playground", "/query"))
	}

	srv := http.Server{
		Addr:    ":" + appConfig.APIPort,
		Handler: router,
	}

	go func() {
		logger.Info("server initialized successfully", slog.String("url", "http://localhost:"+appConfig.APIPort+"/query"))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("failed to start server", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", slog.Any("error", err))
	}

	logger.Info("Server stoped gracefully")
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
	userService := user.NewService(queries, pool)
	profileService := profile.NewService(queries)
	authService := auth.NewService(userService)

	queuePublisher := queue.NewRabbitMQPublisher(rabbitMQCh, cfg.ContentQueueName)

	episodeService := episode.NewService(queries, storageService, profileService, queuePublisher)
	contentService := content.NewService(queries, pool, storageService, profileService, queuePublisher)
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
