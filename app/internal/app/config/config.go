package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Host string
	Port int
}

func NewConfig() (*Config, error) {
	var err error

	configName := "config"
	_ = godotenv.Load()
	if os.Getenv("CONFIG_NAME") != "" {
		configName = os.Getenv("CONFIG_NAME")
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")
	viper.WatchConfig()

	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}

	log.Info("config parsed")

	return cfg, nil
}

type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Bucket    string
}

var MinioClientConfig = MinioConfig{
	Endpoint:  "localhost:9000",
	AccessKey: "minio123",
	SecretKey: "minio123",
	UseSSL:    false,
	Bucket:    "software-images",
}

func NewMinioClient() (*minio.Client, error) {
	minioClient, err := minio.New(MinioClientConfig.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(MinioClientConfig.AccessKey, MinioClientConfig.SecretKey, ""),
		Secure: MinioClientConfig.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	log.Printf("Successfully connected to MinIO at %s", MinioClientConfig.Endpoint)
	return minioClient, nil
}
