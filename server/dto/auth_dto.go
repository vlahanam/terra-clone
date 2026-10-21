package dto

// [VN] RegisterRequest định nghĩa cấu trúc dữ liệu gửi lên khi đăng ký tài khoản mới
type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	FullName    string `json:"full_name" binding:"required"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

// [VN] LoginRequest định nghĩa thông tin đăng nhập với email và password
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// [VN] RefreshTokenRequest định nghĩa request cấp lại Access Token qua Refresh Token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// [VN] AuthResponse định nghĩa cấu trúc dữ liệu trả về sau khi đăng ký hoặc đăng nhập thành công
type AuthResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	TokenType    string        `json:"token_type"`
	ExpiresIn    int64         `json:"expires_in"`
	User         *UserResponse `json:"user"`
}
