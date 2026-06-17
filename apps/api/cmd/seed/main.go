package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/database/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
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
		log.Fatalf("failed to create new db pool: %v\n", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping db: %v\n", err)
	}

	log.Println("Connected to database. Starting seed...")

	queries := sqlc.New(pool)

	users := seedUsers(ctx, queries, pool)
	profiles := seedProfiles(ctx, queries, users)
	seedContents(ctx, queries, profiles)

	log.Println("Seed finished successfully!")
}

func seedUsers(ctx context.Context, queries *sqlc.Queries, pool *pgxpool.Pool) []sqlc.User {
	log.Println("Seeding users...")
	passHash, _ := bcrypt.GenerateFromPassword([]byte("Senha@123"), bcrypt.DefaultCost)

	usersData := []struct {
		email string
		name  string
		cpf   string
		role  int
	}{
		{"admin@netflix.com", "João Administrador", "11122233344", 1},
		{"maria@gmail.com", "Maria Silva", "55566677788", 2},
		{"carlos@hotmail.com", "Carlos Pereira", "99988877766", 2},
	}

	var createdUsers []sqlc.User
	for _, ud := range usersData {
		// Use raw query for role_id update or check if it already exists
		var existingUser sqlc.User
		err := pool.QueryRow(ctx, "SELECT id, email, name, cpf, role_id FROM users WHERE email = $1", ud.email).Scan(
			&existingUser.ID, &existingUser.Email, &existingUser.Name, &existingUser.Cpf, &existingUser.RoleID,
		)
		if err == nil {
			createdUsers = append(createdUsers, existingUser)
			continue
		}

		user, err := queries.CreateUser(ctx, sqlc.CreateUserParams{
			ID:       uuid.New(),
			Email:    ud.email,
			Name:     ud.name,
			Cpf:      ud.cpf,
			Password: string(passHash),
		})
		if err != nil {
			log.Fatalf("failed to create user %s: %v", ud.email, err)
		}

		if ud.role == 1 {
			_, err = pool.Exec(ctx, "UPDATE users SET role_id = 1 WHERE id = $1", user.ID)
			if err != nil {
				log.Fatalf("failed to update user role %s: %v", ud.email, err)
			}
			user.RoleID = 1
		}
		createdUsers = append(createdUsers, user)
	}

	return createdUsers
}

func seedProfiles(ctx context.Context, queries *sqlc.Queries, users []sqlc.User) []sqlc.Profile {
	log.Println("Seeding profiles...")
	var allProfiles []sqlc.Profile

	for _, user := range users {
		existingProfiles, _ := queries.ListProfilesByUser(ctx, user.ID)
		if len(existingProfiles) > 0 {
			allProfiles = append(allProfiles, existingProfiles...)
			continue
		}

		p1, err := queries.CreateProfile(ctx, sqlc.CreateProfileParams{
			ID:                  uuid.New(),
			UserID:              user.ID,
			Name:                "Principal",
			HasParentalControls: false,
		})
		if err != nil {
			log.Fatalf("failed to create profile for user %s: %v", user.Name, err)
		}
		allProfiles = append(allProfiles, p1)

		p2, err := queries.CreateProfile(ctx, sqlc.CreateProfileParams{
			ID:                  uuid.New(),
			UserID:              user.ID,
			Name:                "Infantil",
			HasParentalControls: true,
		})
		if err != nil {
			log.Fatalf("failed to create profile for user %s: %v", user.Name, err)
		}
		allProfiles = append(allProfiles, p2)
	}

	return allProfiles
}

func seedContents(ctx context.Context, queries *sqlc.Queries, profiles []sqlc.Profile) {
	log.Println("Seeding contents...")

	existingContents, _ := queries.ListContents(ctx)
	if len(existingContents) > 0 {
		log.Println("Contents already exist, skipping...")
		return
	}

	movie1ID := uuid.New()
	err := queries.CreateContent(ctx, sqlc.CreateContentParams{
		ID:             movie1ID,
		Title:          "Interestelar",
		ContentType:    sqlc.ContentTypeMOVIE,
		GenreID:        4,
		Description:    "Uma equipe de exploradores viaja através de um buraco de minhoca no espaço na tentativa de garantir a sobrevivência da humanidade.",
		ReleaseDate:    time.Date(2014, 11, 6, 0, 0, 0, 0, time.UTC),
		MaturityRating: sqlc.MaturityRating10,
	})
	if err != nil {
		log.Fatalf("failed to create content: %v", err)
	}
	_, err = queries.CreateMovie(ctx, sqlc.CreateMovieParams{
		ContentID:       movie1ID,
		DurationMinutes: 169,
	})
	if err != nil {
		log.Fatalf("failed to create movie: %v", err)
	}

	movie2ID := uuid.New()
	err = queries.CreateContent(ctx, sqlc.CreateContentParams{
		ID:             movie2ID,
		Title:          "Oppenheimer",
		ContentType:    sqlc.ContentTypeMOVIE,
		GenreID:        3,
		Description:    "A história do cientista americano J. Robert Oppenheimer e seu papel no desenvolvimento da bomba atômica.",
		ReleaseDate:    time.Date(2023, 7, 20, 0, 0, 0, 0, time.UTC),
		MaturityRating: sqlc.MaturityRating16,
	})
	if err != nil {
		log.Fatalf("failed to create content: %v", err)
	}

	_, err = queries.CreateMovie(ctx, sqlc.CreateMovieParams{
		ContentID:       movie2ID,
		DurationMinutes: 180,
	})
	if err != nil {
		log.Fatalf("failed to create movie: %v", err)
	}

	series1ID := uuid.New()
	err = queries.CreateContent(ctx, sqlc.CreateContentParams{
		ID:             series1ID,
		Title:          "Breaking Bad",
		ContentType:    sqlc.ContentTypeSERIES,
		GenreID:        3, // Drama
		Description:    "Um professor de química com câncer terminal se une a um ex-aluno para fabricar e vender metanfetamina para garantir o futuro de sua família.",
		ReleaseDate:    time.Date(2008, 1, 20, 0, 0, 0, 0, time.UTC),
		MaturityRating: sqlc.MaturityRating18,
	})
	if err != nil {
		log.Fatalf("failed to create content: %v", err)
	}
	_, err = queries.CreateSeries(ctx, series1ID)
	if err != nil {
		log.Fatalf("failed to create series: %v", err)
	}

	episodesData := []struct {
		season   int32
		ep       int32
		title    string
		duration int32
	}{
		{1, 1, "Piloto", 58},
		{1, 2, "O Gato na Bolsa", 48},
		{1, 3, "E a Bolsa está no Rio", 48},
	}

	for _, ep := range episodesData {
		_, err = queries.CreateEpisode(ctx, sqlc.CreateEpisodeParams{
			ID:              uuid.New(),
			SeriesID:        series1ID,
			Season:          ep.season,
			EpisodeNumber:   ep.ep,
			Title:           ep.title,
			DurationMinutes: ep.duration,
		})
		if err != nil {
			log.Fatalf("failed to create episode: %v", err)
		}
	}
}
