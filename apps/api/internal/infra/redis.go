package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/config"
	"github.com/gomodule/redigo/redis"
)

func InitializeRedis(cfg *config.Config) (*redis.Pool, error) {
	pool := &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		DialContext: func(context.Context) (redis.Conn, error) {
			return redis.Dial("tcp", cfg.RedisAddr(), redis.DialPassword(cfg.RedisPass))
		},
	}

	conn := pool.Get()
	defer conn.Close()

	if err := conn.Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	if _, err := conn.Do("PING"); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return pool, nil
}
