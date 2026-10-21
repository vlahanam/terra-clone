CREATE TABLE IF NOT EXISTS user_credentials (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    password_hash VARCHAR(255) NOT NULL,
    token_version INT NOT NULL DEFAULT 1,
    last_login_at TIMESTAMPTZ,
    failed_login_attempts INT NOT NULL DEFAULT 0,
    locked_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON COLUMN user_credentials.user_id IS '[VN] ID của người dùng';
COMMENT ON COLUMN user_credentials.password_hash IS '[VN] Mật khẩu người dùng đã được băm an toàn (bcrypt/argon2)';
COMMENT ON COLUMN user_credentials.token_version IS '[VN] Phiên bản token dùng để thu hồi toàn bộ token/session cũ khi đổi mật khẩu hoặc logout all';
COMMENT ON COLUMN user_credentials.last_login_at IS '[VN] Thời điểm đăng nhập thành công gần nhất';
COMMENT ON COLUMN user_credentials.failed_login_attempts IS '[VN] Số lần nhập sai mật khẩu liên tiếp (phòng chống brute-force)';
COMMENT ON COLUMN user_credentials.locked_until IS '[VN] Thời điểm tài khoản tự động được mở khóa sau khi bị khóa tạm thời';
COMMENT ON COLUMN user_credentials.created_at IS '[VN] Thời điểm tạo bản ghi';
COMMENT ON COLUMN user_credentials.updated_at IS '[VN] Thời điểm cập nhật bản ghi gần nhất';
