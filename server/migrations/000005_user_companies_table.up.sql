CREATE TABLE IF NOT EXISTS user_companies (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    public_id VARCHAR(50) UNIQUE NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    job_title VARCHAR(100),
    department VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    permissions JSONB NOT NULL DEFAULT '{}'::jsonb,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    joined_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_user_companies_user_company UNIQUE (user_id, company_id)
);

COMMENT ON TABLE user_companies IS '[VN] Bảng lưu trữ liên kết và quyền hạn giữa người dùng và công ty';
COMMENT ON COLUMN user_companies.id IS '[VN] Khóa chính nội bộ dạng UUIDv7';
COMMENT ON COLUMN user_companies.public_id IS '[VN] Mã định danh công khai của liên kết (tiền tố ucmp_)';
COMMENT ON COLUMN user_companies.user_id IS '[VN] Khóa ngoại tham chiếu đến bảng users(id)';
COMMENT ON COLUMN user_companies.company_id IS '[VN] Khóa ngoại tham chiếu đến bảng companies(id)';
COMMENT ON COLUMN user_companies.role IS '[VN] Vai trò của người dùng trong công ty (owner, admin, member, guest)';
COMMENT ON COLUMN user_companies.job_title IS '[VN] Chức danh công việc của người dùng trong công ty';
COMMENT ON COLUMN user_companies.department IS '[VN] Phòng ban làm việc của người dùng trong công ty';
COMMENT ON COLUMN user_companies.status IS '[VN] Trạng thái hoạt động (active, inactive, pending, suspended, deleted)';
COMMENT ON COLUMN user_companies.permissions IS '[VN] Danh sách quyền hạn tùy biến dạng JSONB';
COMMENT ON COLUMN user_companies.metadata IS '[VN] Dữ liệu cấu hình mở rộng dạng JSONB';
COMMENT ON COLUMN user_companies.joined_at IS '[VN] Thời điểm người dùng chính thức tham gia công ty';
COMMENT ON COLUMN user_companies.created_at IS '[VN] Thời điểm tạo bản ghi';
COMMENT ON COLUMN user_companies.updated_at IS '[VN] Thời điểm cập nhật bản ghi gần nhất';

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_companies_public_id ON user_companies(public_id);
CREATE INDEX IF NOT EXISTS idx_user_companies_user_id ON user_companies(user_id);
CREATE INDEX IF NOT EXISTS idx_user_companies_company_id ON user_companies(company_id);
CREATE INDEX IF NOT EXISTS idx_user_companies_role ON user_companies(role);
CREATE INDEX IF NOT EXISTS idx_user_companies_status ON user_companies(status);
CREATE INDEX IF NOT EXISTS idx_user_companies_permissions ON user_companies USING gin (permissions);
CREATE INDEX IF NOT EXISTS idx_user_companies_metadata ON user_companies USING gin (metadata);
