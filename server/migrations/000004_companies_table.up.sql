CREATE TABLE IF NOT EXISTS companies (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    public_id VARCHAR(50) UNIQUE NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    tax_id VARCHAR(50),
    email VARCHAR(255),
    phone VARCHAR(50),
    address TEXT,
    website VARCHAR(255),
    logo_url TEXT,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    country_code VARCHAR(10),
    timezone VARCHAR(100),
    currency VARCHAR(10),
    legal_name VARCHAR(255),
    city VARCHAR(100),
    state_or_province VARCHAR(100),
    postal_code VARCHAR(20),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON COLUMN companies.id IS '[VN] ID nội bộ của công ty (UUIDv7)';
COMMENT ON COLUMN companies.public_id IS '[VN] Mã định danh công khai của công ty (tiền tố cmp_)';
COMMENT ON COLUMN companies.code IS '[VN] Mã công ty duy nhất định danh tổ chức (viết hoa)';
COMMENT ON COLUMN companies.name IS '[VN] Tên pháp lý hoặc tên thương mại của công ty';
COMMENT ON COLUMN companies.tax_id IS '[VN] Mã số thuế doanh nghiệp';
COMMENT ON COLUMN companies.email IS '[VN] Địa chỉ email liên hệ chính thức của công ty';
COMMENT ON COLUMN companies.phone IS '[VN] Số điện thoại hotline hoặc liên hệ của công ty';
COMMENT ON COLUMN companies.address IS '[VN] Địa chỉ trụ sở chính hoặc văn phòng công ty';
COMMENT ON COLUMN companies.website IS '[VN] Trang thông tin điện tử website chính thức của công ty';
COMMENT ON COLUMN companies.logo_url IS '[VN] Đường dẫn URL ảnh logo công ty';
COMMENT ON COLUMN companies.description IS '[VN] Mô tả tóm tắt hoặc giới thiệu về công ty';
COMMENT ON COLUMN companies.status IS '[VN] Trạng thái hoạt động: active, inactive, suspended, deleted';
COMMENT ON COLUMN companies.owner_id IS '[VN] ID người dùng sở hữu hoặc đại diện quản trị công ty';
COMMENT ON COLUMN companies.country_code IS '[VN] Mã quốc gia đăng ký kinh doanh theo chuẩn ISO 3166-1 alpha-2';
COMMENT ON COLUMN companies.timezone IS '[VN] Múi giờ hoạt động chính của công ty (chuẩn IANA)';
COMMENT ON COLUMN companies.currency IS '[VN] Đơn vị tiền tệ chính của doanh nghiệp theo chuẩn ISO 4217 (ví dụ: VND, USD)';
COMMENT ON COLUMN companies.legal_name IS '[VN] Tên pháp lý đầy đủ trên giấy phép kinh doanh quốc tế';
COMMENT ON COLUMN companies.city IS '[VN] Thành phố đặt trụ sở chính của công ty';
COMMENT ON COLUMN companies.state_or_province IS '[VN] Tỉnh hoặc Bang nơi công ty đặt trụ sở';
COMMENT ON COLUMN companies.postal_code IS '[VN] Mã bưu chính của trụ sở công ty';
COMMENT ON COLUMN companies.metadata IS '[VN] Cấu hình và dữ liệu mở rộng dạng JSONB';
COMMENT ON COLUMN companies.created_at IS '[VN] Thời điểm tạo bản ghi công ty';
COMMENT ON COLUMN companies.updated_at IS '[VN] Thời điểm cập nhật bản ghi công ty gần nhất';

CREATE UNIQUE INDEX IF NOT EXISTS idx_companies_public_id ON companies(public_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_companies_code ON companies(code);
CREATE INDEX IF NOT EXISTS idx_companies_owner_id ON companies(owner_id);
CREATE INDEX IF NOT EXISTS idx_companies_status ON companies(status);
CREATE INDEX IF NOT EXISTS idx_companies_tax_id ON companies(tax_id);
CREATE INDEX IF NOT EXISTS idx_companies_metadata ON companies USING gin (metadata);
CREATE INDEX IF NOT EXISTS idx_companies_country_code ON companies(country_code);
