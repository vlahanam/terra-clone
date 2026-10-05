package internal

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBConfig *DBConfig
}

func NewConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Println("[VN] Warning: .env file not found, relying on system environment variables")
	}

	dbConfig := &DBConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Name:     os.Getenv("DB_NAME"),
		Username: os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
	}

	if dbConfig.Host == "" || dbConfig.Name == "" {
		return nil, fmt.Errorf("[VN] missing required database environment variables")
	}

	return &Config{
		DBConfig: dbConfig,
	}, nil
}