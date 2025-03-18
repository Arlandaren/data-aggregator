package api

import (
	"context"
	"github.com/Arlandaren/pgxWrappy/pkg/postgres"
	"service/internal/domains/api/models"
	"service/internal/infrastructure/storage/minio"
	"service/internal/infrastructure/storage/redis"
)

type Repository struct {
	db  *postgres.Wrapper
	rdb *redis.RDB
	s3  *minio.Minio
}

func NewRepository(db *postgres.Wrapper, rdb *redis.RDB, s3 *minio.Minio) *Repository {
	return &Repository{
		db:  db,
		rdb: rdb,
		s3:  s3,
	}
}

func (r *Repository) UploadFIle(ctx context.Context, file models.FileUpload) (string, error) {
	link, err := r.s3.UploadFileToMinio(ctx, file.BucketName, file.ObjectName, file.FileSize, file.FileBytes)
	if err != nil {
		return "", err
	}
	return link, nil
}
