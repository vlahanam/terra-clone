package internal

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// [VN] DBConfig chứa các tham số cấu hình kết nối PostgreSQL
type DBConfig struct {
	Host     string
	Port     string
	Name     string
	Username string
	Password string
	Sslmode  string
}

// [VN] DSN tạo chuỗi Data Source Name cho kết nối PostgreSQL
func (c *DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		c.Host, c.Username, c.Password, c.Name, c.Port, c.Sslmode,
	)
}

// [VN] ConnectDb thực hiện kết nối tới database và kiểm tra ping theo nguyên tắc Fail-Fast
func (c *DBConfig) ConnectDb() (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(c.DSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB instance: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}