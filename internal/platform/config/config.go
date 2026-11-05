package config

import (
	"errors"
	"fmt"
	"os"
)

type S3Config struct {
	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSSessionToken    string
	AWSBucket          string
	AWSEndpoint        string
}

type Config struct {
	Host                  string
	Port                  string
	DatabaseURL           string
	LogLevel              string
	AuthAPIURL            string
	AuthJWTSecret         string
	AuthServiceEmail      string
	AuthServicePassword   string
	PaymentsEnabled       bool
	PaymentsAPIURL        string
	PaymentsAPIKey        string
	PaymentsWebhookSecret string
	S3Region              string
	S3AccessKeyID         string
	S3SecretAccessKey     string
	S3Bucket              string
	S3Endpoint            string
	S3                    S3Config
}

func Load() (Config, error) {
	cfg := Config{
		Host:                  getenv("VAMO_BACKEND_HOST", "0.0.0.0"),
		Port:                  getenv("VAMO_BACKEND_PORT", "8080"),
		DatabaseURL:           os.Getenv("DATABASE_URL"),
		LogLevel:              getenv("LOG_LEVEL", "info"),
		AuthAPIURL:            getenv("AUTH_API_URL", "http://localhost:8082"),
		AuthJWTSecret:         os.Getenv("AUTH_JWT_SECRET"),
		AuthServiceEmail:      os.Getenv("AUTH_SERVICE_EMAIL"),
		AuthServicePassword:   os.Getenv("AUTH_SERVICE_PASSWORD"),
		PaymentsEnabled:       getenv("PAYMENTS_ENABLED", "false") == "true",
		PaymentsAPIURL:        getenv("PAYMENTS_API_URL", "http://localhost:8008/api/v1/payments/preferences"),
		PaymentsAPIKey:        os.Getenv("PAYMENTS_API_KEY"),
		PaymentsWebhookSecret: os.Getenv("PAYMENTS_WEBHOOK_SECRET"),
		S3Region:              os.Getenv("S3_REGION"),
		S3AccessKeyID:         os.Getenv("S3_ACCESS_KEY_ID"),
		S3SecretAccessKey:     os.Getenv("S3_SECRET_ACCESS_KEY"),
		S3Bucket:              os.Getenv("S3_BUCKET"),
		S3Endpoint:            os.Getenv("S3_ENDPOINT"),
		S3: S3Config{
			AWSRegion:          os.Getenv("S3_REGION"),
			AWSAccessKeyID:     os.Getenv("S3_ACCESS_KEY_ID"),
			AWSSecretAccessKey: os.Getenv("S3_SECRET_ACCESS_KEY"),
			AWSSessionToken:    os.Getenv("S3_SESSION_TOKEN"),
			AWSBucket:          os.Getenv("S3_BUCKET"),
			AWSEndpoint:        os.Getenv("S3_ENDPOINT"),
		},
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	return cfg, nil
}

func (c Config) Address() string { return fmt.Sprintf("%s:%s", c.Host, c.Port) }

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
