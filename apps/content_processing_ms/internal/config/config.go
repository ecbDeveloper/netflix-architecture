package config

import (
	"fmt"
	"log/slog"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	S3Region         string `env:"S3_REGION,notEmpty"`
	S3AccessKeyID    string `env:"S3_ACCESS_KEY_ID,notEmpty"`
	S3SecretKey      string `env:"S3_SECRET_ACCESS_KEY,notEmpty"`
	S3EndpointURL    string `env:"S3_ENDPOINT_URL,notEmpty"`
	S3BucketName     string `env:"S3_BUCKET_NAME,notEmpty"`
	RabbitMQHost     string `env:"RABBITMQ_HOST,notEmpty"`
	RabbitMQPort     string `env:"RABBITMQ_PORT,notEmpty"`
	RabbitMQUser     string `env:"RABBITMQ_USER,notEmpty"`
	RabbitMQPass     string `env:"RABBITMQ_PASS,notEmpty"`
	ContentQueueName string `env:"CONTENT_QUEUE_NAME,notEmpty"`
	DBName           string `env:"DB_NAME,notEmpty"`
	DBUser           string `env:"DB_USER,notEmpty"`
	DBPass           string `env:"DB_PASS,notEmpty"`
	DBPort           string `env:"DB_PORT,notEmpty"`
	DBHost           string `env:"DB_HOST,notEmpty"`
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
