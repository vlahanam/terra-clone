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

func NewConfig() (*config, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Println("[VN] Warning: .env file not found, relying on system environment variables")
	}

	appConfig := &AppConfig{
		Port: os.Getenv("PORT"),
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
		return nil, fmt.Errorf("[VN] missing required database environment variables")
	}

	return &config{
		AppConfig: appConfig,
		DBConfig:  dbConfig,
	}, nil
}
