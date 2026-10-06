package internal

import (
	"fmt"
	"log"
)

func Run() {
	cf, err := NewConfig()
	if err != nil {
		log.Fatalf("[VN] Lỗi khởi tạo cấu hình: %v", err)
	}

	db, err := cf.DBConfig.ConnectDb()
	if err != nil {
		log.Fatalf("[VN] Lỗi kết nối database: %v", err)
	}
	_ = db

	r := SetupRouter(cf.AppConfig)

	port := fmt.Sprintf(":%s", cf.AppConfig.Port)
	if err := r.Run(port); err != nil {
		log.Fatalf("[VN]Server bị dừng đột ngột: %v", err)
	}
}