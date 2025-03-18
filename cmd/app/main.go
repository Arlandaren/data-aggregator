package main

import (
	wrapper "github.com/Arlandaren/pgxWrappy/pkg/postgres"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"service/internal/application"
	"service/internal/infrastructure/config"
	"service/internal/infrastructure/logger"
	"service/internal/infrastructure/storage/minio"
	"service/internal/infrastructure/storage/postgres"
	"service/internal/infrastructure/storage/redis"
)

func main() {
	logger.Init()

	cfg := config.NewConfig()

	dbPool, err := postgres.InitPostgres(cfg.Postgres, 5)
	if err != nil {
		log.Fatalf("failed to initialize Postgres: %v", err)
	}
	db := wrapper.NewWrapper(dbPool)
	rdb, err := redis.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatalf("failed to initialize Redis: %v", err)
	}
	s3, err := minio.NewMinio(cfg.Minio)
	if err != nil {
		log.Fatalf("failed to initialize Minio: %v", err)
	}
	app := application.NewApp(db, rdb, s3, gin.Default(), cfg)

	app.Run()
}
