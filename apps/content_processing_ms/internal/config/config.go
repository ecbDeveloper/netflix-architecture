package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Env              string
	S3Region         string
	S3AccessKeyID    string
	S3SecretKey      string
	S3EndpointURL    string
	S3BucketName     string
	RabbitMQHost     string
	RabbitMQPort     string
	RabbitMQUser     string
	RabbitMQPass     string
	ContentQueueName string
	DBName           string
	DBUser           string
	DBPass           string
	DBPort           string
	DBHost           string
}

func LoadConfig(logger *slog.Logger) *Config {
	err := godotenv.Load()
	if err != nil {
		logger.Info("No .env file found, using OS environment variables")
	}

	return &Config{
		Env:              os.Getenv("ENV"),
		DBName:           os.Getenv("DB_NAME"),
		DBUser:           os.Getenv("DB_USER"),
		DBPass:           os.Getenv("DB_PASS"),
		DBPort:           os.Getenv("DB_PORT"),
		DBHost:           os.Getenv("DB_HOST"),
		S3Region:         os.Getenv("S3_REGION"),
		S3AccessKeyID:    os.Getenv("S3_ACCESS_KEY_ID"),
		S3SecretKey:      os.Getenv("S3_SECRET_ACCESS_KEY"),
		S3EndpointURL:    os.Getenv("S3_ENDPOINT_URL"),
		S3BucketName:     os.Getenv("S3_BUCKET_NAME"),
		RabbitMQHost:     os.Getenv("RABBITMQ_HOST"),
		RabbitMQPort:     os.Getenv("RABBITMQ_PORT"),
		RabbitMQUser:     os.Getenv("RABBITMQ_USER"),
		RabbitMQPass:     os.Getenv("RABBITMQ_PASS"),
		ContentQueueName: os.Getenv("CONTENT_QUEUE_NAME"),
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPass, c.DBName,
	)
}
