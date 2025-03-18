package minio

import (
	"context"
	"service/internal/infrastructure/config"
	"testing"
)

func TestMinioConn(t *testing.T) {
	cfg := &config.MinioConfig{
		Endpoint:        "localhost:9000",
		AccessKeyID:     "Owner",
		SecretAccessKey: "123456789",
		UseSSL:          false,
	}
	minio, err := NewMinio(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	_, err = minio.Client.ListBuckets(ctx)
	if err != nil {
		t.Fatalf("Не удалось выполнить ListBuckets: %v", err)
	}
}
