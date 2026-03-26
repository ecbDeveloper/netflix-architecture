package resolvers

import "github.com/vektah/gqlparser/v2/gqlerror"

var (
	ErrInvalidEmailOrPassword = gqlerror.Errorf("invalid credentials")
	ErrUserNotLoggedIn        = gqlerror.Errorf("user must be logged in")
)
