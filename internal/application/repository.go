package application

import (
	"github.com/Arlandaren/pgxWrappy/pkg/postgres"
	"service/internal/domains/api"
	"service/internal/domains/person"
	"service/internal/infrastructure/storage/minio"
	"service/internal/infrastructure/storage/redis"
)

type Repository struct {
	Person *person.Repository
	Api    *api.Repository
}

func NewRepository(db *postgres.Wrapper, rdb *redis.RDB, s3 *minio.Minio) *Repository {
	return &Repository{
		Person: person.NewRepository(db, rdb),
		Api:    api.NewRepository(db, rdb, s3),
	}
}
