package config

import (
	"fmt"
	"os"
)

type Config struct {
	APIPort                string
	Env                    string
	HistoryGRPCAddr        string
	RecommendationGRPCAddr string
	DBHost                 string
	DBPort                 string
	DBUser                 string
	DBPass                 string
	DBName                 string
	RedisHost              string
	RedisPort              string
	RedisPass              string
	S3Region               string
	S3AccessKeyID          string
	S3SecretAccessKey      string
	S3EndPointURL          string
	S3BucketName           string
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

func Load() (*Config, error) {
	cfg := &Config{
		APIPort:                os.Getenv("API_PORT"),
		Env:                    os.Getenv("ENV"),
		HistoryGRPCAddr:        os.Getenv("HISTORY_GRPC_ADDR"),
		RecommendationGRPCAddr: os.Getenv("RECOMMENDATION_GRPC_ADDR"),
		DBHost:                 os.Getenv("DB_HOST"),
		DBPort:                 os.Getenv("DB_PORT"),
		DBUser:                 os.Getenv("DB_USER"),
		DBPass:                 os.Getenv("DB_PASS"),
		DBName:                 os.Getenv("DB_NAME"),
		RedisHost:              os.Getenv("REDIS_HOST"),
		RedisPort:              os.Getenv("REDIS_PORT"),
		RedisPass:              os.Getenv("REDIS_PASS"),
		S3Region:               os.Getenv("S3_REGION"),
		S3AccessKeyID:          os.Getenv("S3_ACCESS_KEY_ID"),
		S3SecretAccessKey:      os.Getenv("S3_SECRET_ACCESS_KEY"),
		S3EndPointURL:          os.Getenv("S3_ENDPOINT_URL"),
		S3BucketName:           os.Getenv("S3_BUCKET_NAME"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	required := map[string]string{
		"API_PORT":                 c.APIPort,
		"ENV":                      c.Env,
		"HISTORY_GRPC_ADDR":        c.HistoryGRPCAddr,
		"RECOMMENDATION_GRPC_ADDR": c.RecommendationGRPCAddr,
		"DB_HOST":                  c.DBHost,
		"DB_PORT":                  c.DBPort,
		"DB_USER":                  c.DBUser,
		"DB_PASS":                  c.DBPass,
		"DB_NAME":                  c.DBName,
		"REDIS_HOST":               c.RedisHost,
		"REDIS_PORT":               c.RedisPort,
		"REDIS_PASS":               c.RedisPass,
		"S3_REGION":                c.S3Region,
		"S3_ACCESS_KEY_ID":         c.S3AccessKeyID,
		"S3_SECRET_ACCESS_KEY":     c.S3SecretAccessKey,
		"S3_ENDPOINT_URL":          c.S3EndPointURL,
		"S3_BUCKET_NAME":           c.S3BucketName,
	}

	for key, val := range required {
		if val == "" {
			return fmt.Errorf("required environment variable %s is not set", key)
		}
	}

	return nil
}
