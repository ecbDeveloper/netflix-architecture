package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/ecbDeveloper/netflix-architecture/internal/database/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	ctx := context.Background()

	loggerHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(loggerHandler)

	err := godotenv.Load()
	if err != nil {
		logger.Error("failed to load .env file", slog.String("error", err.Error()))
		os.Exit(1)
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_NAME"),
	)

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		logger.Error("failed to initialize db pool to seed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	tx, err := pool.Begin(ctx)
	if err != nil {
		logger.Error("failed to begin transaction", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer tx.Rollback(ctx)

	qtx := sqlc.New(pool).WithTx(tx)

	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPass := os.Getenv("ADMIN_PASSWORD")

	if adminEmail == "" || adminPass == "" {
		logger.Info("ADMIN_EMAIL or ADMIN_PASSWORD not set")
		os.Exit(1)
	}

	_, err = qtx.GetUserByEmail(ctx, adminEmail)
	if err == nil {
		logger.Info("Admin user already exists, skipping user creation")
	} else {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
		if err != nil {
			logger.Error("failed to hash admin password", slog.String("error", err.Error()))
			os.Exit(1)
		}

		adminID := uuid.New()
		_, err = qtx.CreateUser(ctx, sqlc.CreateUserParams{
			ID:       adminID,
			Email:    adminEmail,
			Name:     "Admin User",
			Cpf:      "00000000000",
			Password: string(hashedPassword),
		})
		if err != nil {
			logger.Error("failed to create admin user", slog.String("error", err.Error()))
			os.Exit(1)
		}

		_, err = tx.Exec(ctx, "UPDATE users SET role = 1 WHERE id = $1", adminID)
		if err != nil {
			logger.Error("failed to set user as admin", slog.String("error", err.Error()))
			os.Exit(1)
		}

		_, err = qtx.CreateProfile(ctx, sqlc.CreateProfileParams{
			ID:                  uuid.New(),
			UserID:              adminID,
			Name:                "Admin Profile",
			HasParentalControls: false,
		})
		if err != nil {
			logger.Error("failed to create admin profile", slog.String("error", err.Error()))
		}

	}

	movieID := uuid.New()
	_, err = qtx.CreateMovie(ctx, sqlc.CreateMovieParams{
		ID:              movieID,
		Title:           "The Go Gopher Movie",
		Description:     "A movie about the Go Gopher adventuring in the cloud.",
		DurationMinutes: 120,
		ReleaseDate:     pgtype.Date{Time: time.Now(), Valid: true},
		MaturityRating:  "13",
		ContentUrl:      "https://example.com/movies/gopher.mp4",
	})
	if err != nil {
		logger.Error("failed to seed movie", slog.String("error", err.Error()))
	}

	serie, err := qtx.CreateSerie(ctx, sqlc.CreateSerieParams{
		Title:          "Breaking Bugs",
		Description:    "A programmer turned bug hunter.",
		ReleaseDate:    pgtype.Date{Time: time.Now(), Valid: true},
		MaturityRating: "TV-MA",
	})
	if err != nil {
		logger.Error("failed to seed series", slog.String("error", err.Error()))
	}

	episodeID := uuid.New()
	_, err = qtx.CreateEpisode(ctx, sqlc.CreateEpisodeParams{
		ID:              episodeID,
		SeriesID:        serie.ID,
		Season:          1,
		EpisodeNumber:   1,
		Title:           "Pilot: Panic at the Disco",
		DurationMinutes: 45,
	})
	if err != nil {
		logger.Error("failed to seed episode", slog.String("error", err.Error()))
	}

	if err := tx.Commit(ctx); err != nil {
		logger.Error("failed to commit transaction", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("Seeding completed!")
}
