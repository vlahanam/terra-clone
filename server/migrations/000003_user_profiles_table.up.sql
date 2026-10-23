CREATE TABLE IF NOT EXISTS user_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    date_of_birth DATE,
    gender VARCHAR(20),
    phone_number VARCHAR(20),
    address TEXT,
    country_code VARCHAR(10),
    nationality VARCHAR(100),
    timezone VARCHAR(100),
    locale VARCHAR(20),
    city VARCHAR(100),
    state_or_province VARCHAR(100),
    postal_code VARCHAR(20),
    preferences JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON COLUMN user_profiles.user_id IS '[VN] ID của người dùng';
COMMENT ON COLUMN user_profiles.date_of_birth IS '[VN] Ngày sinh của người dùng';
COMMENT ON COLUMN user_profiles.gender IS '[VN] Giới tính của người dùng';
COMMENT ON COLUMN user_profiles.phone_number IS '[VN] Số điện thoại liên lạc của người dùng';
COMMENT ON COLUMN user_profiles.address IS '[VN] Địa chỉ chi tiết của người dùng';
COMMENT ON COLUMN user_profiles.country_code IS '[VN] Mã quốc gia cư trú theo chuẩn ISO 3166-1 alpha-2 (ví dụ: VN, US)';
COMMENT ON COLUMN user_profiles.nationality IS '[VN] Quốc tịch của người dùng (ví dụ: Vietnamese, American)';
COMMENT ON COLUMN user_profiles.timezone IS '[VN] Múi giờ IANA của người dùng (ví dụ: Asia/Ho_Chi_Minh, UTC)';
COMMENT ON COLUMN user_profiles.locale IS '[VN] Ngôn ngữ và định dạng địa phương theo BCP 47 (ví dụ: vi-VN, en-US)';
COMMENT ON COLUMN user_profiles.city IS '[VN] Thành phố nơi cư trú';
COMMENT ON COLUMN user_profiles.state_or_province IS '[VN] Tỉnh, Bang hoặc Vùng lãnh thổ nơi cư trú';
COMMENT ON COLUMN user_profiles.postal_code IS '[VN] Mã bưu chính hoặc Zip code';
COMMENT ON COLUMN user_profiles.preferences IS '[VN] Cài đặt và tùy chọn cá nhân hóa dưới định dạng JSONB';
COMMENT ON COLUMN user_profiles.created_at IS '[VN] Thời điểm bản ghi được tạo';
COMMENT ON COLUMN user_profiles.updated_at IS '[VN] Thời điểm bản ghi được cập nhật gần nhất';

-- [VN] Tạo index GIN cho cột preferences
CREATE INDEX idx_user_profiles_preferences ON user_profiles USING gin (preferences);
-- [VN] Tạo index GIN cho cột country_code, nếu đã tồn tại thì bỏ qua
CREATE INDEX IF NOT EXISTS idx_user_profiles_country_code ON user_profiles(country_code);