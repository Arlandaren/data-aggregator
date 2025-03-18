package config

import (
	"errors"
	"os"
	"strconv"
)

type Config struct {
	Postgres *PostgresConfig
	Address  *Address
	Redis    *RedisConfig
	Minio    *MinioConfig
	Env      string
}

func NewConfig() *Config {
	pg, err := GetPostgres()
	if err != nil {
		panic(err)
	}
	mn, err := GetMinio()
	if err != nil {
		panic(err)
	}
	return &Config{
		Postgres: pg,
		Address:  GetAddress(),
		Redis:    GetRedis(),
		Env:      GetEnvironment(),
		Minio:    mn,
	}
}

func GetPostgres() (*PostgresConfig, error) {
	pgConn := os.Getenv("PG_STRING")
	if pgConn == "" {
		return nil, errors.New("not found PG_STRING")
	}
	return &PostgresConfig{
		ConnStr: pgConn,
	}, nil
}

func GetAddress() *Address {
	httpAddress := os.Getenv("HTTP_ADDRESS")

	if httpAddress == "" {
		httpAddress = ":8086"
	}

	return &Address{
		Http: httpAddress,
	}
}

func GetEnvironment() string {
	return os.Getenv("ENVIRONMENT")
}

func GetRedis() *RedisConfig {
	redisAddress := os.Getenv("REDIS_ADDRESS")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	return &RedisConfig{
		Addr:     redisAddress,
		Password: redisPassword,
	}
}

func GetTogether() *TogetherAIConfig {
	api := os.Getenv("TOGETHER_API")
	url := os.Getenv("TOGETHER_URL")
	return &TogetherAIConfig{
		APIKey: api,
		URL:    url,
	}
}

func GetMinio() (*MinioConfig, error) {
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	minioAccessKey := os.Getenv("MINIO_ROOT_USER")
	minioSecretKey := os.Getenv("MINIO_ROOT_PASSWORD")

	minioSsl, err := strconv.ParseBool(os.Getenv("MINIO_SSL"))
	if err != nil {
		return nil, err
	}

	return &MinioConfig{
		Endpoint:        minioEndpoint,
		AccessKeyID:     minioAccessKey,
		SecretAccessKey: minioSecretKey,
		UseSSL:          minioSsl,
	}, nil
}
