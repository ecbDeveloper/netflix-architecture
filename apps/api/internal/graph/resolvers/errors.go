package resolvers

import (
	"errors"
	"fmt"

	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/apperror"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidEmailOrPassword = gqlerror.Errorf("invalid credentials")
	ErrUserNotLoggedIn        = gqlerror.Errorf("user must be logged in")
)

func handleError(err error) *gqlerror.Error {
	var validationErr *apperror.ValidationError
	if errors.As(err, &validationErr) {
		return &gqlerror.Error{
			Message: validationErr.Message,
			Extensions: map[string]any{
				"code": "BAD_REQUEST",
			},
		}
	}

	var conflictErr *apperror.ConflictError
	if errors.As(err, &conflictErr) {
		return &gqlerror.Error{
			Message: fmt.Sprintf("%s is already in use", conflictErr.Field),
			Extensions: map[string]any{
				"code": "CONFLICT",
			},
		}
	}

	var notFoundErr *apperror.NotFoundError
	if errors.As(err, &notFoundErr) {
		return &gqlerror.Error{
			Message: fmt.Sprintf("%s not found", notFoundErr.Entity),
			Extensions: map[string]any{
				"code": "NOT_FOUND",
			},
		}
	}

	var forbiddenErr *apperror.ForbiddenError
	if errors.As(err, &forbiddenErr) {
		return &gqlerror.Error{
			Message: forbiddenErr.Message,
			Extensions: map[string]any{
				"code": "FORBIDDEN",
			},
		}
	}

	var unauthorizedErr *apperror.UnauthorizedError
	if errors.As(err, &unauthorizedErr) {
		return &gqlerror.Error{
			Message: unauthorizedErr.Message,
			Extensions: map[string]any{
				"code": "UNAUTHORIZED",
			},
		}
	}

	return &gqlerror.Error{
		Message: "internal error, try again later",
		Extensions: map[string]any{
			"code": "INTERNAL_SERVER",
		},
	}
}

func handleGRPCError(err error) *gqlerror.Error {
	st, ok := status.FromError(err)
	if !ok {
		return handleError(err)
	}

	switch st.Code() {
	case codes.InvalidArgument:
		return &gqlerror.Error{
			Message: st.Message(),
			Extensions: map[string]any{"code": "BAD_REQUEST"},
		}
	case codes.NotFound:
		return &gqlerror.Error{
			Message: st.Message(),
			Extensions: map[string]any{"code": "NOT_FOUND"},
		}
	case codes.Unauthenticated:
		return &gqlerror.Error{
			Message: st.Message(),
			Extensions: map[string]any{"code": "UNAUTHORIZED"},
		}
	case codes.PermissionDenied:
		return &gqlerror.Error{
			Message: st.Message(),
			Extensions: map[string]any{"code": "FORBIDDEN"},
		}
	default:
		return &gqlerror.Error{
			Message: "internal error from external service",
			Extensions: map[string]any{"code": "INTERNAL_SERVER"},
		}
	}
}

