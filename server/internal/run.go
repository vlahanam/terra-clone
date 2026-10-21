package internal

import (
	"fmt"
	"log/slog"
)

// [VN] Run khởi chạy toàn bộ ứng dụng: nạp cấu hình, cấu hình logger, kết nối database và khởi động Gin router
func Run() {
	// [VN] Nạp cấu hình ứng dụng
	cf, err := NewConfig()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		return
	}

	// [VN] Khởi tạo logger ghi file và access log
	accessWriter, cleanupLogger, err := InitLogger(cf.AppConfig)
	if err != nil {
		fmt.Printf("failed to initialize log roller: %v\n", err)
		return
	}
	defer func() {
		if err := cleanupLogger(); err != nil {
			fmt.Printf("failed to cleanup log roller: %v\n", err)
		}
	}()

	// [VN] Khởi tạo kết nối cơ sở dữ liệu PostgreSQL
	db, err := cf.DBConfig.ConnectDb()
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		return
	}

	// [VN] Khởi tạo router và lắng nghe HTTP requests
	r := SetupRouter(cf.AppConfig, db, accessWriter)
	port := fmt.Sprintf(":%s", cf.AppConfig.Port)
	if err := r.Run(port); err != nil {
		slog.Error("server stopped unexpectedly", "error", err)
	}
}
