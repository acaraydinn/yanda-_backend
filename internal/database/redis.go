package database

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/yandas/backend/internal/config"
)

// ConnectRedis establishes a connection to Redis
func ConnectRedis(cfg *config.Config) (*redis.Client, error) {
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx := context.Background()
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	log.Println("✅ Redis connected successfully")
	return client, nil
}
