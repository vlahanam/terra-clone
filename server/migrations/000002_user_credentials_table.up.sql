CREATE TABLE IF NOT EXISTS user_credentials (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    username VARCHAR(255),
    password_hash VARCHAR(255),
    auth_provider VARCHAR(50) NOT NULL DEFAULT 'local',
    provider_user_id VARCHAR(255),
    token_version INT NOT NULL DEFAULT 1,
    last_login_at TIMESTAMPTZ,
    failed_login_attempts INT NOT NULL DEFAULT 0,
    locked_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_local_password CHECK (
        auth_provider != 'local' OR (username IS NOT NULL AND password_hash IS NOT NULL)
    )
);

COMMENT ON COLUMN user_credentials.user_id IS '[VN] ID của người dùng';
COMMENT ON COLUMN user_credentials.username IS '[VN] Tên đăng nhập của tài khoản';
COMMENT ON COLUMN user_credentials.password_hash IS '[VN] Mật khẩu người dùng null nếu đăng nhập hoàn toàn qua Social Login';
COMMENT ON COLUMN user_credentials.auth_provider IS '[VN] Phương thức xác thực: local, google';
COMMENT ON COLUMN user_credentials.provider_user_id IS '[VN] Mã định danh duy nhất của user từ bên cung cấp dịch vụ OAuth';
COMMENT ON COLUMN user_credentials.token_version IS '[VN] Phiên bản token dùng để thu hồi toàn bộ token/session cũ khi đổi mật khẩu hoặc logout all';
COMMENT ON COLUMN user_credentials.last_login_at IS '[VN] Thời điểm đăng nhập thành công gần nhất';
COMMENT ON COLUMN user_credentials.failed_login_attempts IS '[VN] Số lần nhập sai mật khẩu liên tiếp (phòng chống brute-force)';
COMMENT ON COLUMN user_credentials.locked_until IS '[VN] Thời điểm tài khoản tự động được mở khóa sau khi bị khóa tạm thời';
COMMENT ON COLUMN user_credentials.created_at IS '[VN] Thời điểm tạo bản ghi';
COMMENT ON COLUMN user_credentials.updated_at IS '[VN] Thời điểm cập nhật bản ghi gần nhất';

-- [VN] Index hỗ trợ tìm kiếm qua provider nếu cần đăng nhập bằng mạng xã hội
CREATE INDEX idx_user_credentials_provider ON user_credentials(auth_provider, provider_user_id);
-- [VN] Index hỗ trợ tìm kiếm nhanh khi đăng nhập bằng username
CREATE INDEX idx_user_credentials_username ON user_credentials(username);