package internal

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type config struct {
	AppConfig *AppConfig
	DBConfig  *DBConfig
}

// [VN] NewConfig nạp cấu hình từ file .env hoặc biến môi trường hệ thống theo nguyên tắc Fail-Fast
func NewConfig() (*config, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found, relying on system environment variables")
	}

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "local"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		if appEnv == "staging" || appEnv == "production" {
			return nil, fmt.Errorf("missing required environment variable JWT_SECRET for environment %s", appEnv)
		}
		jwtSecret = "terra_dev_jwt_secret_key_2026_very_secure_string_min_32_bytes"
	}

	appConfig := &AppConfig{
		AppEnv:    appEnv,
		Port:      os.Getenv("PORT"),
		JWTSecret: jwtSecret,
	}

	dbConfig := &DBConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Name:     os.Getenv("DB_NAME"),
		Username: os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Sslmode:  os.Getenv("DB_SSLMODE"),
	}

	if appConfig.Port == "" || dbConfig.Host == "" || dbConfig.Name == "" {
		return nil, fmt.Errorf("missing required database environment variables")
	}

	return &config{
		AppConfig: appConfig,
		DBConfig:  dbConfig,
	}, nil
}
