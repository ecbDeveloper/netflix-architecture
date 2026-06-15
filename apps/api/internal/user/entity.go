package user

import "github.com/google/uuid"

type User struct {
	ID     uuid.UUID
	Name   string
	Email  Email
	CPF    CPF
	RoleID int32
}

func (u User) IsAdmin() bool { return u.RoleID == 1 }

func (u User) IsMember() bool { return u.RoleID == 2 }
