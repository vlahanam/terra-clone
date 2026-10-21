package internal

// [VN] AppConfig chứa các tham số cấu hình cấp ứng dụng
type AppConfig struct {
	AppEnv string
	Port string
	JWTSecret string
}
