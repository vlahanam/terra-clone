package internal

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBConfig struct {
	Host     string
	Port     string
	Name     string
	Username string
	Password string
	Sslmode  string
}

func (c *DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		c.Host, c.Username, c.Password, c.Name, c.Port, c.Sslmode,
	)
}

func (c *DBConfig) ConnectDb() (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(c.DSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("[VN] Không thể mở kết nối database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("[VN] Không thể lấy instance sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("[VN] Không thể ping tới database: %w", err)
	}

	return db, nil
}