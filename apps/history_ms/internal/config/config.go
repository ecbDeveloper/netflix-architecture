package config

import (
	"fmt"
	"log/slog"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	DBName   string `env:"DB_NAME,notEmpty"`
	DBUser   string `env:"DB_USER,notEmpty"`
	DBPass   string `env:"DB_PASS,notEmpty"`
	DBPort   string `env:"DB_PORT,notEmpty"`
	DBHost   string `env:"DB_HOST,notEmpty"`
	GRPCPort string `env:"GRPC_PORT,notEmpty"`
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found, loading config from environment variables")
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPass, c.DBName,
	)
}
