package common

import "time"

const (
	// [VN] Tiền tố tạo publish_id
	PREFIX_USER    = "usr"
	PREFIX_COMPANY = "cmp"

	// [VN] Các loại token JWT và thời hạn hiệu lực
	JWT_TOKEN_TYPE_ACCESS      = "access"
	JWT_TOKEN_TYPE_REFRESH     = "refresh"
	JWT_ACCESS_TOKEN_DURATION  = 15 * time.Minute
	JWT_REFRESH_TOKEN_DURATION = 7 * 24 * time.Hour

	// [VN] Các hằng số quản lý chống tấn công Brute-force
	ACC_MAX_LOGIN_ATTEMPTS = 5                // [VN] Số lần đăng nhập sai tối đa liên tiếp cho phép
	ACC_LOCK_TIME          = 15 * time.Minute // [VN] Thời gian tạm khóa tài khoản sau khi vượt quá số lần nhập sai
)
