package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"service/internal/infrastructure/config"
)

type RDB struct {
	Client *redis.Client
}

func NewRedisClient(config *config.RedisConfig) (*RDB, error) {

	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       0,
	})

	if err := client.Ping(context.TODO()).Err(); err != nil {
		return &RDB{}, err
	}

	return &RDB{
		Client: client,
	}, nil
}
