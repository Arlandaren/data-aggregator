package minio

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
	"service/internal/infrastructure/config"
	"time"

	"bytes"
)

type Minio struct {
	Client  *minio.Client
	cfg     *config.MinioConfig
	Allowed map[string]bool
}

func NewMinio(cfg *config.MinioConfig) (*Minio, error) {
	endpoint := cfg.Endpoint
	accessKeyID := cfg.AccessKeyID
	secretAccessKey := cfg.SecretAccessKey
	useSSL := cfg.UseSSL

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	allowed := map[string]bool{
		".gif": true,
		".png": true,
		".jpg": true,
	}

	log.Println("MinIO клиент успешно инициализирован.")
	return &Minio{Client: minioClient, cfg: cfg, Allowed: allowed}, nil
}

func (m *Minio) UploadFileToMinio(ctx context.Context, bucketName, objectName string, fileSize int64, fileBytes []byte) (string, error) {

	exists, errBucketExists := m.Client.BucketExists(ctx, bucketName)
	if errBucketExists != nil {
		return "", fmt.Errorf("ошибка проверки существования бакета: %v", errBucketExists)
	}
	if !exists {
		err := m.Client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return "", fmt.Errorf("не удалось создать бакет: %v", err)
		}
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {"AWS": ["*"]},
					"Action": ["s3:GetObject"],
					"Resource": ["arn:aws:s3:::%s/*"]
				}
			]
		}`, bucketName)

		if err = m.Client.SetBucketPolicy(ctx, bucketName, policy); err != nil {
			return "", fmt.Errorf("ошибка установки политики бакета: %v", err)
		}
	}

	file := bytes.NewReader(fileBytes)

	uploadInfo, err := m.Client.PutObject(ctx, bucketName, objectName, file, fileSize, minio.PutObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("ошибка загрузки файла: %v", err)
	}

	fmt.Printf("Файл успешно загружен. Информация: %+v\n", uploadInfo)

	publicURL := fmt.Sprintf("http://%s/%s/%s", m.cfg.Endpoint, bucketName, objectName)

	return publicURL, nil
}

func (m *Minio) DownloadFileFromMinio(bucketName, objectName, downloadPath string) error {
	ctx := context.Background()

	err := m.Client.FGetObject(ctx, bucketName, objectName, downloadPath, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("ошибка скачивания файла: %v", err)
	}

	fmt.Printf("Файл успешно скачан по пути: %s\n", downloadPath)
	return nil
}

func (m *Minio) GenerateResignedURL(objectName string) (string, error) {
	ctx := context.Background()

	bucketName := "images"
	urlExpiry := 768 * time.Hour

	url, err := m.Client.PresignedGetObject(ctx, bucketName, objectName, urlExpiry, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка генерации ссылки для %s: %v", objectName, err)
	}
	return url.String(), nil
}
