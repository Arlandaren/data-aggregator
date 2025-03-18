package redis

import (
	"service/internal/infrastructure/config"
	"testing"
)

func TestRedisConn(t *testing.T) {
	cfg := &config.RedisConfig{
		Addr:     "localhost:6378",
		Password: "1234",
	}
	rdb, err := NewRedisClient(cfg)
	if err != nil {
		t.Fatalf("failed to connect to Redis: %v", err)
	}

	if rdb.Client == nil {
		t.Fatal("Redis client is nil")
	}
}
