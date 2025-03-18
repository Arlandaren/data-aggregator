package api

import (
	"context"
	"io"
	"mime/multipart"
	"path/filepath"
	"service/internal/domains/api/models"
	"service/internal/infrastructure/storage/models/dto"
	"strings"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) UploadFile(ctx context.Context, fileHeader *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))

	if !s.repo.s3.Allowed[ext] {
		return "", dto.ErrWrongFile
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	req := models.FileUpload{
		FileSize:   fileHeader.Size,
		FileBytes:  fileBytes,
		ObjectName: fileHeader.Filename,
		BucketName: "images",
	}

	fileLink, err := s.repo.UploadFIle(ctx, req)
	if err != nil {
		return "", err
	}
	return fileLink, nil
}
