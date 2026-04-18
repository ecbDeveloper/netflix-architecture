package content

import (
	"context"
	"fmt"

	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/graph/model"
	"github.com/google/uuid"
)

type Service interface {
	CreateContent(ctx context.Context, input model.CreateContentInput) (*model.Content, error)
	UpdateContent(ctx context.Context, id uuid.UUID, input model.UpdateContentInput) (*model.Content, error)
	DeleteContent(ctx context.Context, id uuid.UUID) (bool, error)
	ListContents(ctx context.Context) ([]*model.Content, error)
	ListKidsContents(ctx context.Context) ([]*model.Content, error)
	ListContentsByType(ctx context.Context, contentType model.ContentType) ([]*model.Content, error)
	ListContentsByGenre(ctx context.Context, genreID int32) ([]*model.Content, error)
}

type ServiceImpl struct {
	queries *sqlc.Queries
}

func NewService(queries *sqlc.Queries) Service {
	return &ServiceImpl{queries: queries}
}

func (s *ServiceImpl) CreateContent(ctx context.Context, input model.CreateContentInput) (*model.Content, error) {
	panic(fmt.Errorf("not implemented: CreateContent - createContent"))
}

// UpdateContent is the resolver for the updateContent field.
func (s *ServiceImpl) UpdateContent(ctx context.Context, id uuid.UUID, input model.UpdateContentInput) (*model.Content, error) {
	panic(fmt.Errorf("not implemented: UpdateContent - updateContent"))
}

// DeleteContent is the resolver for the deleteContent field.
func (s *ServiceImpl) DeleteContent(ctx context.Context, id uuid.UUID) (bool, error) {
	panic(fmt.Errorf("not implemented: DeleteContent - deleteContent"))
}

// GetContent is the resolver for the getContent field.
func (s *ServiceImpl) GetContent(ctx context.Context, id uuid.UUID) (*model.Content, error) {
	panic(fmt.Errorf("not implemented: GetContent - getContent"))
}

// ListContents is the resolver for the listContents field.
func (s *ServiceImpl) ListContents(ctx context.Context) ([]*model.Content, error) {
	panic(fmt.Errorf("not implemented: ListContents - listContents"))
}

// ListKidsContents is the resolver for the listKidsContents field.
func (s *ServiceImpl) ListKidsContents(ctx context.Context) ([]*model.Content, error) {
	panic(fmt.Errorf("not implemented: ListKidsContents - listKidsContents"))
}

// ListContentsByType is the resolver for the listContentsByType field.
func (s *ServiceImpl) ListContentsByType(ctx context.Context, contentType model.ContentType) ([]*model.Content, error) {
	panic(fmt.Errorf("not implemented: ListContentsByType - listContentsByType"))
}

// ListContentsByGenre is the resolver for the listContentsByGenre field.
func (s *ServiceImpl) ListContentsByGenre(ctx context.Context, genreID int32) ([]*model.Content, error) {
	panic(fmt.Errorf("not implemented: ListContentsByGenre - listContentsByGenre"))
}
