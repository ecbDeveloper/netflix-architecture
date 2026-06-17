package content

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/apperror"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/database/sqlc"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/graph/model"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/infra/storage"
	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/profile"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service interface {
	CreateContent(ctx context.Context, input model.CreateContentInput) (uuid.UUID, error)
	UpdateContent(ctx context.Context, id uuid.UUID, input model.UpdateContentInput) (*model.Content, error)
	DeleteContent(ctx context.Context, id uuid.UUID) error
	GetContent(ctx context.Context, id uuid.UUID, profileID uuid.UUID, userID uuid.UUID) (*model.Content, error)
	ListContents(ctx context.Context, profileID uuid.UUID, userID uuid.UUID) ([]*model.Content, error)
	ListContentsByType(ctx context.Context, profileID uuid.UUID, userID uuid.UUID, contentType model.ContentType) ([]*model.Content, error)
	ListContentsByGenre(ctx context.Context, profileID uuid.UUID, userID uuid.UUID, genreID int32) ([]*model.Content, error)
}

type ServiceImpl struct {
	repo           Repository
	queries        *sqlc.Queries
	pool           *pgxpool.Pool
	storage        storage.Service
	profileService profile.Service
}

func NewService(queries *sqlc.Queries, pool *pgxpool.Pool, storage storage.Service, ps profile.Service) Service {
	return &ServiceImpl{
		repo:           queries,
		queries:        queries,
		pool:           pool,
		storage:        storage,
		profileService: ps,
	}
}

func (s *ServiceImpl) CreateContent(ctx context.Context, input model.CreateContentInput) (uuid.UUID, error) {
	if strings.TrimSpace(input.Title) == "" {
		return uuid.Nil, &apperror.ValidationError{Field: "title", Message: "title is required"}
	}

	if strings.TrimSpace(input.Description) == "" {
		return uuid.Nil, &apperror.ValidationError{Field: "description", Message: "description is required"}
	}

	if strings.TrimSpace(input.ReleaseDate) == "" {
		return uuid.Nil, &apperror.ValidationError{Field: "releaseDate", Message: "releaseDate is required"}
	}

	timeLayout := "2006-01-02"
	date, err := time.Parse(timeLayout, input.ReleaseDate)
	if err != nil {
		return uuid.Nil, &apperror.ValidationError{Field: "releaseDate", Message: "releaseDate invalid format"}
	}

	if strings.TrimSpace(string(input.MaturityRating)) == "" {
		return uuid.Nil, &apperror.ValidationError{Field: "maturityRating", Message: "maturityRating is required"}
	}

	if input.GenreID <= 0 {
		return uuid.Nil, &apperror.ValidationError{Field: "genreId", Message: "genreId is required"}
	}

	if input.ContentType == model.ContentTypeMovie {
		if input.ContentFile.File == nil {
			return uuid.Nil, &apperror.ValidationError{Field: "contentFile", Message: "contentFile is required to movies"}
		}

		if input.DurationMinutes == nil {
			return uuid.Nil, &apperror.ValidationError{Field: "durationMinutes", Message: "durationMinutes is required to movies"}
		}
	}

	genres, err := s.repo.ListContentGenres(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get content genres list from database: %w", err)
	}

	genreWasFounded := false
	genreIsFamily := false
	for _, genre := range genres {
		if input.GenreID == genre.ID {
			genreWasFounded = true
			genreIsFamily = genre.Description == "Family"
			break
		}
	}

	if !genreWasFounded {
		return uuid.Nil, &apperror.ValidationError{Field: "genreId", Message: "invalid genreId"}
	}

	if input.MaturityRating == model.MaturityRatingL && !genreIsFamily {
		return uuid.Nil, &apperror.UnprocessableEntityError{Message: "invalid genre to a children's content"}
	}

	if strings.TrimSpace(string(input.ContentType)) == "" {
		return uuid.Nil, &apperror.ValidationError{Field: "contentType", Message: "contentType is required"}
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to initialize transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)
	contentID := uuid.New()

	createContentParams := sqlc.CreateContentParams{
		ID:             contentID,
		ContentType:    sqlc.ContentType(input.ContentType),
		Title:          input.Title,
		Description:    input.Description,
		ReleaseDate:    date,
		MaturityRating: graphQLToDBMaturityRating(input.MaturityRating),
		GenreID:        input.GenreID,
	}

	err = qtx.CreateContent(ctx, createContentParams)
	if err != nil {
		if apperror.IsUniqueViolation(err) {
			field := apperror.UniqueViolationField(err)
			return uuid.Nil, &apperror.ConflictError{Field: field}
		}

		return uuid.Nil, fmt.Errorf("failed to insert content on database: %w", err)
	}

	var fileKey string
	if input.ContentType == model.ContentTypeMovie {
		fileKey = fmt.Sprintf("raw/%s.mp4", contentID.String())
		err = s.storage.Upload(ctx, contentID, input.ContentFile.File)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to upload content file: %w", err)
		}

		createMovieParams := sqlc.CreateMovieParams{
			ContentID:       contentID,
			DurationMinutes: *input.DurationMinutes,
		}

		_, err = qtx.CreateMovie(ctx, createMovieParams)
		if err != nil {
			if err := s.storage.DeleteFile(context.Background(), fileKey); err != nil {
				return uuid.Nil, fmt.Errorf("failed to delete movie file: %w", err)
			}

			return uuid.Nil, fmt.Errorf("failed to insert movie on database: %w", err)
		}
	}

	if input.ContentType == model.ContentTypeSeries {
		_, err = qtx.CreateSeries(ctx, contentID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to insert series on database: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		if input.ContentType == model.ContentTypeMovie {
			if err := s.storage.DeleteFile(context.Background(), fileKey); err != nil {
				return uuid.Nil, fmt.Errorf("failed to delete movie file: %w", err)
			}
		}

		return uuid.Nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return contentID, nil
}

func (s *ServiceImpl) UpdateContent(ctx context.Context, id uuid.UUID, input model.UpdateContentInput) (*model.Content, error) {
	current, err := s.repo.GetContent(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "content"}
		}
		return nil, fmt.Errorf("failed to get content from database: %w", err)
	}

	if input.Title != nil && strings.TrimSpace(*input.Title) == "" {
		return nil, &apperror.ValidationError{Field: "title", Message: "title is required"}
	}

	if input.Description != nil && strings.TrimSpace(*input.Description) == "" {
		return nil, &apperror.ValidationError{Field: "description", Message: "description is required"}
	}

	if input.ReleaseDate != nil && strings.TrimSpace(*input.ReleaseDate) == "" {
		return nil, &apperror.ValidationError{Field: "releaseDate", Message: "releaseDate is required"}
	}

	timeLayout := "2006-01-02"
	var date time.Time
	if input.ReleaseDate != nil {
		date, err = time.Parse(timeLayout, *input.ReleaseDate)
		if err != nil {
			return nil, &apperror.ValidationError{Field: "releaseDate", Message: "releaseDate invalid format"}
		}
	}

	if input.MaturityRating != nil && strings.TrimSpace(string(*input.MaturityRating)) == "" {
		return nil, &apperror.ValidationError{Field: "maturityRating", Message: "maturityRating is required"}
	}

	if input.GenreID != nil && *input.GenreID <= 0 {
		return nil, &apperror.ValidationError{Field: "genreId", Message: "genreId must be greater than zero"}
	}

	if input.GenreID != nil {
		genreWasFounded := false
		genres, err := s.repo.ListContentGenres(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get content genres list from database: %w", err)
		}

		for _, genre := range genres {
			if *input.GenreID == genre.ID {
				genreWasFounded = true
				break
			}
		}

		if !genreWasFounded {
			return nil, &apperror.ValidationError{Field: "genreId", Message: "invalid genreId"}
		}
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	updateContentParams := sqlc.UpdateContentParams{
		ID:             id,
		Title:          current.Title,
		Description:    current.Description,
		ReleaseDate:    current.ReleaseDate,
		MaturityRating: current.MaturityRating,
		GenreID:        current.GenreID,
	}

	if input.Title != nil {
		updateContentParams.Title = *input.Title
	}

	if input.Description != nil {
		updateContentParams.Description = *input.Description
	}

	if input.ReleaseDate != nil {
		updateContentParams.ReleaseDate = date
	}

	if input.MaturityRating != nil {
		updateContentParams.MaturityRating = graphQLToDBMaturityRating(*input.MaturityRating)
	}

	if input.GenreID != nil {
		updateContentParams.GenreID = *input.GenreID
	}

	content, err := qtx.UpdateContent(ctx, updateContentParams)
	if err != nil {
		if apperror.IsUniqueViolation(err) {
			field := apperror.UniqueViolationField(err)
			return nil, &apperror.ConflictError{Field: field}
		}
		return nil, fmt.Errorf("failed to update content on database: %w", err)
	}

	var currentMovie sqlc.GetMovieRow
	var oldURL string
	if current.ContentType == sqlc.ContentTypeMOVIE {
		currentMovie, err = s.repo.GetMovie(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, &apperror.NotFoundError{Entity: "movie"}
			}
			return nil, fmt.Errorf("failed to get movie from database: %w", err)
		}

		updateMovieParams := sqlc.UpdateMovieParams{
			ContentID:       id,
			DurationMinutes: currentMovie.DurationMinutes,
			ContentUrl:      currentMovie.ContentUrl,
			Status:          currentMovie.Status,
		}

		if input.ContentFile.File != nil {
			err = s.storage.Upload(ctx, id, input.ContentFile.File)
			if err != nil {
				return nil, fmt.Errorf("failed to upload content file: %w", err)
			}
			oldURL = currentMovie.ContentUrl.String
			updateMovieParams.Status = sqlc.ContentStatusPENDING
		}

		if input.DurationMinutes != nil {
			updateMovieParams.DurationMinutes = *input.DurationMinutes
		}

		_, err = qtx.UpdateMovie(ctx, updateMovieParams)
		if err != nil {
			if input.ContentFile != nil && input.ContentFile.File != nil {
				go s.storage.DeleteFile(context.Background(), oldURL)
			}

			return nil, fmt.Errorf("failed to update movie on database: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		if current.ContentType == sqlc.ContentTypeMOVIE {
			fileKey := fmt.Sprintf("raw/%s.mp4", id.String())
			go s.storage.DeleteFile(context.Background(), fileKey)
		}

		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	if input.ContentFile != nil && input.ContentFile.File != nil && oldURL != "" {
		go s.storage.DeleteFile(context.Background(), oldURL)
	}

	entity := toContentEntity(content)
	if entity.IsMovie() {
		return toGraphQlModel(content, pgTextToStringPtr(currentMovie.ContentUrl), &currentMovie.DurationMinutes), nil
	}

	return toGraphQlModel(content, nil, nil), nil
}

func (s *ServiceImpl) DeleteContent(ctx context.Context, id uuid.UUID) error {
	content, err := s.repo.GetContent(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &apperror.NotFoundError{Entity: "content"}
		}
		return fmt.Errorf("failed to get content from database: %w", err)
	}

	contentEntity := toContentEntity(content)

	var movie sqlc.GetMovieRow
	if contentEntity.IsMovie() {
		movie, err = s.repo.GetMovie(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return &apperror.NotFoundError{Entity: "movie"}
			}
			return fmt.Errorf("failed to get movie from database: %w", err)
		}
	}

	var episodes []sqlc.Episode
	if contentEntity.IsSeries() {
		episodes, err = s.repo.ListEpisodesBySeries(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return &apperror.NotFoundError{Entity: "series"}
			}
			return fmt.Errorf("failed to list episodes from database: %w", err)
		}
	}

	err = s.repo.DeleteContent(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete content from database: %w", err)
	}

	if contentEntity.IsMovie() {
		if movie.ContentUrl.Valid && movie.ContentUrl.String != "" {
			go s.storage.DeleteFile(context.Background(), movie.ContentUrl.String)
		}
	}

	if contentEntity.IsSeries() {
		for _, episode := range episodes {
			if episode.ContentUrl.Valid && episode.ContentUrl.String != "" {
				go s.storage.DeleteFile(context.Background(), episode.ContentUrl.String)
			}
		}
	}

	return nil
}

func (s *ServiceImpl) ListContents(ctx context.Context, profileID uuid.UUID, userID uuid.UUID) ([]*model.Content, error) {
	profile, err := s.profileService.GetProfile(ctx, profileID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile %v: %w", profileID, err)
	}

	var contents []sqlc.Content
	if profile.HasParentalControls {
		contents, err = s.repo.ListKidsContents(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list contents from database: %w", err)
		}
	} else {
		contents, err = s.repo.ListContents(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list contents from database: %w", err)
		}
	}

	result := make([]*model.Content, len(contents))
	for i, c := range contents {
		entity := toContentEntity(c)
		if entity.IsMovie() {
			movie, err := s.repo.GetMovie(ctx, c.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get movie from database: %w", err)
			}
			result[i] = toGraphQlModel(c, pgTextToStringPtr(movie.ContentUrl), &movie.DurationMinutes)
		}
		if entity.IsSeries() {
			result[i] = toGraphQlModel(c, nil, nil)
		}
	}

	return result, nil
}

func (s *ServiceImpl) ListContentsByType(ctx context.Context, profileID uuid.UUID, userID uuid.UUID, contentType model.ContentType) ([]*model.Content, error) {
	profile, err := s.profileService.GetProfile(ctx, profileID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile %v: %w", profileID, err)
	}

	var contents []sqlc.Content
	if profile.HasParentalControls {
		contents, err = s.repo.ListKidsContentsByType(ctx, sqlc.ContentType(contentType))
		if err != nil {
			return nil, fmt.Errorf("failed to list contents by type from database: %w", err)
		}
	} else {
		contents, err = s.repo.ListContentsByType(ctx, sqlc.ContentType(contentType))
		if err != nil {
			return nil, fmt.Errorf("failed to list contents by type from database: %w", err)
		}
	}

	result := make([]*model.Content, len(contents))
	for i, c := range contents {
		entity := toContentEntity(c)
		if entity.IsMovie() {
			movie, err := s.repo.GetMovie(ctx, c.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get movie from database: %w", err)
			}
			result[i] = toGraphQlModel(c, pgTextToStringPtr(movie.ContentUrl), &movie.DurationMinutes)
		}
		if entity.IsSeries() {
			result[i] = toGraphQlModel(c, nil, nil)
		}
	}

	return result, nil
}

func (s *ServiceImpl) ListContentsByGenre(ctx context.Context, profileID uuid.UUID, userID uuid.UUID, genreID int32) ([]*model.Content, error) {
	genres, err := s.repo.ListContentGenres(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get content genres list from database: %w", err)
	}

	genreWasFounded := false
	for _, genre := range genres {
		if genreID == genre.ID {
			genreWasFounded = true
			break
		}
	}

	if !genreWasFounded {
		return nil, &apperror.ValidationError{Field: "genreId", Message: "invalid genreId"}
	}

	profile, err := s.profileService.GetProfile(ctx, profileID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile %v: %w", profileID, err)
	}

	var contents []sqlc.Content
	if profile.HasParentalControls {
		contents, err = s.repo.ListKidsContentsByGenre(ctx, genreID)
		if err != nil {
			return nil, fmt.Errorf("failed to list contents by type from database: %w", err)
		}
	} else {
		contents, err = s.repo.ListContentsByGenre(ctx, genreID)
		if err != nil {
			return nil, fmt.Errorf("failed to list contents by type from database: %w", err)
		}
	}

	result := make([]*model.Content, len(contents))
	for i, c := range contents {
		entity := toContentEntity(c)
		if entity.IsMovie() {
			movie, err := s.repo.GetMovie(ctx, c.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get movie from database: %w", err)
			}
			result[i] = toGraphQlModel(c, pgTextToStringPtr(movie.ContentUrl), &movie.DurationMinutes)
		}
		if entity.IsSeries() {
			result[i] = toGraphQlModel(c, nil, nil)
		}
	}

	return result, nil
}

func (s *ServiceImpl) GetContent(ctx context.Context, id uuid.UUID, profileID uuid.UUID, userID uuid.UUID) (*model.Content, error) {
	profile, err := s.profileService.GetProfile(ctx, profileID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile %v: %w", profileID, err)
	}

	content, err := s.repo.GetContent(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperror.NotFoundError{Entity: "content"}
		}
		return nil, fmt.Errorf("failed to fetch content %v from database: %w", id, err)
	}

	contentEntity := toContentEntity(content)
	if !contentEntity.IsAccessibleBy(profile.HasParentalControls) {
		return nil, &apperror.ForbiddenError{Message: "you can't access this content, because your profile has parental control"}
	}

	if contentEntity.IsMovie() {
		movie, err := s.repo.GetMovie(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, &apperror.NotFoundError{Entity: "movie"}
			}
			return nil, fmt.Errorf("failed to fetch movie %v from database: %w", id, err)
		}
		return toGraphQlModel(content, pgTextToStringPtr(movie.ContentUrl), &movie.DurationMinutes), nil
	}

	return toGraphQlModel(content, nil, nil), nil
}
