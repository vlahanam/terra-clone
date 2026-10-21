package internal

import (
	"fmt"
	"log/slog"
)

func Run() {
	cf, err := NewConfig()
	if err != nil {
		slog.Error("[VN] Lỗi khởi tạo cấu hình", "error", err)
		return
	}

	accessWriter, cleanupLogger, err := InitLogger(cf.AppConfig)
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
		slog.Error("[VN] Lỗi kết nối database", "error", err)
		return
	}

	r := SetupRouter(cf.AppConfig, db, accessWriter)
	port := fmt.Sprintf(":%s", cf.AppConfig.Port)
	if err := r.Run(port); err != nil {
		slog.Error("[VN] Server bị dừng đột ngột", "error", err)
	}
}
