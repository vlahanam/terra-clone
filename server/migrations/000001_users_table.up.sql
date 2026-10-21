CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    public_id VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    avatar_url TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON COLUMN users.id IS '[VN] ID nội bộ của người dùng (UUIDv7)';
COMMENT ON COLUMN users.public_id IS '[VN] Mã định danh công khai của người dùng, hiển thị cho người dùng';
COMMENT ON COLUMN users.email IS '[VN] Địa chỉ email của người dùng';
COMMENT ON COLUMN users.full_name IS '[VN] Họ và tên đầy đủ của người dùng';
COMMENT ON COLUMN users.avatar_url IS '[VN] Đường dẫn ảnh đại diện của người dùng';
COMMENT ON COLUMN users.status IS '[VN] Trạng thái tài khoản: active, inactive, deleted';
COMMENT ON COLUMN users.created_at IS '[VN] Thời điểm tạo bản ghi';
COMMENT ON COLUMN users.updated_at IS '[VN] Thời điểm cập nhật bản ghi gần nhất';

CREATE UNIQUE INDEX idx_users_public_id ON users(public_id);
CREATE INDEX idx_users_email ON users(email);
