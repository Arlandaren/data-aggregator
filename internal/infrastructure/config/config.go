package config

type PostgresConfig struct {
	ConnStr string
}

type RedisConfig struct {
	Addr     string
	Password string
}

type Address struct {
	Http string
}

type TogetherAIConfig struct {
	APIKey string
	URL    string
}

type MinioConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
}
