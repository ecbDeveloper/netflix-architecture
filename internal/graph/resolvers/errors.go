package resolvers

import (
	"errors"

	"github.com/ecbDeveloper/netflix-architecture/internal/apperror"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

var (
	ErrInvalidEmailOrPassword = gqlerror.Errorf("invalid credentials")
	ErrUserNotLoggedIn        = gqlerror.Errorf("user must be logged in")
	ErrGenericInternal        = gqlerror.Errorf("internal error, try again later")
)

func handleError(err error) *gqlerror.Error {
	var validationErr *apperror.ValidationError
	if errors.As(err, &validationErr) {
		return gqlerror.Errorf("%s", validationErr.Message)
	}

	var conflictErr *apperror.ConflictError
	if errors.As(err, &conflictErr) {
		return gqlerror.Errorf("%s is already in use", conflictErr.Field)
	}

	var notFoundErr *apperror.NotFoundError
	if errors.As(err, &notFoundErr) {
		return gqlerror.Errorf("%s not found", notFoundErr.Entity)
	}

	return ErrGenericInternal
}
