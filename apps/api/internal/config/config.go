package config

import (
	"fmt"
	"log/slog"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	APIPort                  string `env:"API_PORT,notEmpty"`
	Env                      string `env:"ENV,notEmpty"`
	HistoryGRPCAddr          string `env:"HISTORY_GRPC_ADDR,notEmpty"`
	RecommendationGRPCAddr   string `env:"RECOMMENDATION_GRPC_ADDR,notEmpty"`
	DBHost                   string `env:"DB_HOST,notEmpty"`
	DBPort                   string `env:"DB_PORT,notEmpty"`
	DBUser                   string `env:"DB_USER,notEmpty"`
	DBPass                   string `env:"DB_PASS,notEmpty"`
	DBName                   string `env:"DB_NAME,notEmpty"`
	RedisHost                string `env:"REDIS_HOST,notEmpty"`
	RedisPort                string `env:"REDIS_PORT,notEmpty"`
	RedisPass                string `env:"REDIS_PASS,notEmpty"`
	S3Region                 string `env:"S3_REGION,notEmpty"`
	S3AccessKeyID            string `env:"S3_ACCESS_KEY_ID,notEmpty"`
	S3SecretAccessKey        string `env:"S3_SECRET_ACCESS_KEY,notEmpty"`
	S3EndPointURL            string `env:"S3_ENDPOINT_URL,notEmpty"`
	S3BucketName             string `env:"S3_BUCKET_NAME,notEmpty"`
	RabbitMQHost             string `env:"RABBITMQ_HOST,notEmpty"`
	RabbitMQPort             string `env:"RABBITMQ_PORT,notEmpty"`
	RabbitMQUser             string `env:"RABBITMQ_USER,notEmpty"`
	RabbitMQPass             string `env:"RABBITMQ_PASS,notEmpty"`
	ContentQueueName         string `env:"CONTENT_QUEUE_NAME,notEmpty"`
	OTELExporterOTLPEndpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT,notEmpty"`
	OTELServiceName          string `env:"OTEL_SERVICE_NAME,notEmpty"`
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

func (c *Config) RedisAddr() string {
	return c.RedisHost + ":" + c.RedisPort
}

func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}
