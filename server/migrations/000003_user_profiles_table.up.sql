CREATE TABLE IF NOT EXISTS user_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    date_of_birth DATE,
    gender VARCHAR(20),
    phone_number VARCHAR(20),
    address TEXT,
    preferences JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON COLUMN user_profiles.user_id IS '[VN] ID của người dùng';
COMMENT ON COLUMN user_profiles.date_of_birth IS 'Ngày sinh của người dùng';
COMMENT ON COLUMN user_profiles.gender IS 'Giới tính của người dùng';
COMMENT ON COLUMN user_profiles.phone_number IS 'Số điện thoại liên lạc của người dùng';
COMMENT ON COLUMN user_profiles.address IS 'Địa chỉ chi tiết của người dùng';
COMMENT ON COLUMN user_profiles.preferences IS 'Cài đặt và tùy chọn cá nhân hóa dưới định dạng JSONB';
COMMENT ON COLUMN user_profiles.created_at IS 'Thời điểm bản ghi được tạo';
COMMENT ON COLUMN user_profiles.updated_at IS 'Thời điểm bản ghi được cập nhật gần nhất';

-- Tạo index GIN cho cột preferences
CREATE INDEX idx_user_profiles_preferences ON user_profiles USING gin (preferences);