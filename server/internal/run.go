package internal

import (
	"fmt"
	"log/slog"
)

func Run() {
	cf, err := NewConfig()
	if err != nil {
		slog.Error("[VN] Lỗi khởi tạo cấu hình: %v", err)
	}

	cleanupLogger, err := InitLogger(cf.AppConfig)
	if err != nil {
		fmt.Printf("[VN] Không thể khởi tạo log roller: %v\n", err)
		return
	}
	defer func() {
		if err := cleanupLogger(); err != nil {
			fmt.Printf("[VN] Lỗi khi đóng log roller: %v\n", err)
		}
	}()

	db, err := cf.DBConfig.ConnectDb()
	if err != nil {
		slog.Error("[VN] Lỗi kết nối database: %v", err)
	}
	_ = db

	r := SetupRouter(cf.AppConfig)
	port := fmt.Sprintf(":%s", cf.AppConfig.Port)
	if err := r.Run(port); err != nil {
		slog.Error("[VN] Server bị dừng đột ngột: %v", err)
	}
}