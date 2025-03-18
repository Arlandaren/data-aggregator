package postgres

import (
	"context"
	"service/internal/infrastructure/config"
	"testing"
)

func TestPostgreConn(t *testing.T) {
	cfg := &config.PostgresConfig{
		ConnStr: "postgres://Owner:1234@localhost:5434/service?sslmode=disable",
	}
	db, err := InitPostgres(cfg, 5)
	if err != nil {
		t.Fatalf("failed to connect to Redis: %v", err)
	}

	if db.Ping(context.TODO()) != nil {
		t.Fatal("Postgre client is nil")
	}
}
