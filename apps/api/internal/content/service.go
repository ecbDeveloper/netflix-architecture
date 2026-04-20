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

func (s *ServiceImpl) UpdateContent(ctx context.Context, id uuid.UUID, input model.UpdateContentInput) (*model.Content, error) {
	panic(fmt.Errorf("not implemented: UpdateContent - updateContent"))
}

func (s *ServiceImpl) DeleteContent(ctx context.Context, id uuid.UUID) (bool, error) {
	panic(fmt.Errorf("not implemented: DeleteContent - deleteContent"))
}

func (s *ServiceImpl) GetContent(ctx context.Context, id uuid.UUID) (*model.Content, error) {
	panic(fmt.Errorf("not implemented: GetContent - getContent"))
}

func (s *ServiceImpl) ListContents(ctx context.Context) ([]*model.Content, error) {
	panic(fmt.Errorf("not implemented: ListContents - listContents"))
}

func (s *ServiceImpl) ListKidsContents(ctx context.Context) ([]*model.Content, error) {
	panic(fmt.Errorf("not implemented: ListKidsContents - listKidsContents"))
}

func (s *ServiceImpl) ListContentsByType(ctx context.Context, contentType model.ContentType) ([]*model.Content, error) {
	panic(fmt.Errorf("not implemented: ListContentsByType - listContentsByType"))
}

func (s *ServiceImpl) ListContentsByGenre(ctx context.Context, genreID int32) ([]*model.Content, error) {
	panic(fmt.Errorf("not implemented: ListContentsByGenre - listContentsByGenre"))
}
