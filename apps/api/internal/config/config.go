package config

import (
	"fmt"
	"os"
)

type Config struct {
	APIPort                string
	Env                    string
	UploadPath             string
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
		UploadPath:             os.Getenv("UPLOAD_PATH"),
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
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	required := map[string]string{
		"API_PORT":                c.APIPort,
		"ENV":                     c.Env,
		"UPLOAD_PATH":             c.UploadPath,
		"HISTORY_GRPC_ADDR":       c.HistoryGRPCAddr,
		"RECOMMENDATION_GRPC_ADDR": c.RecommendationGRPCAddr,
		"DB_HOST":                 c.DBHost,
		"DB_PORT":                 c.DBPort,
		"DB_USER":                 c.DBUser,
		"DB_PASS":                 c.DBPass,
		"DB_NAME":                 c.DBName,
		"REDIS_HOST":              c.RedisHost,
		"REDIS_PORT":              c.RedisPort,
		"REDIS_PASS":              c.RedisPass,
	}

	for key, val := range required {
		if val == "" {
			return fmt.Errorf("required environment variable %s is not set", key)
		}
	}

	return nil
}
