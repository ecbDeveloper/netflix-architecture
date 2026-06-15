package user

import (
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/graph/model"
)

func toGraphQLModel(u sqlc.User) *model.User {
	return &model.User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Cpf:       u.Cpf,
		RoleID:    u.RoleID,
		CreatedAt: u.CreatedAt.String(),
		UpdatedAt: u.UpdatedAt.String(),
	}
}

func toUserEntity(u sqlc.User) User {
	return User{
		ID:     u.ID,
		Name:   u.Name,
		Email:  Email(u.Email),
		CPF:    CPF(u.Cpf),
		RoleID: u.RoleID,
	}
}
