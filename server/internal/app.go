package internal

// [VN] AppConfig chứa các tham số cấu hình cấp ứng dụng
type AppConfig struct {
	// [VN] Môi trường hoạt động của ứng dụng (local, staging, production)
	AppEnv string
	// [VN] Cổng lắng nghe của ứng dụng
	Port string
	// [VN] Chuỗi bí mật dùng để ký và xác thực token JWT
	JWTSecret string
}
