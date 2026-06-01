package app

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/alexedwards/scs/v2"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/config"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/graph"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/graph/model"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/graph/resolvers"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/shared"
	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

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
