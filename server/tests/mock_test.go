package tests

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vlahanam/terra-clone/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// [VN] MockUserRepository giả lập interface UserRepository phục vụ kiểm thử đơn vị cho Service
type MockUserRepository struct {
	mock.Mock
}

// [VN] Create giả lập thao tác lưu người dùng mới vào CSDL
func (m *MockUserRepository) Create(ctx context.Context, tx *gorm.DB, user *models.User) error {
	args := m.Called(ctx, tx, user)
	return args.Error(0)
}

// [VN] FindByID giả lập tìm kiếm người dùng theo khóa chính UUID
func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// [VN] FindByPublicID giả lập tìm kiếm người dùng theo mã định danh công khai
func (m *MockUserRepository) FindByPublicID(ctx context.Context, publicID string) (*models.User, error) {
	args := m.Called(ctx, publicID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// [VN] FindByEmail giả lập tra cứu người dùng qua địa chỉ email
func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// [VN] FindByIDOrPublicID giả lập tìm kiếm người dùng linh hoạt theo UUID hoặc Public ID
func (m *MockUserRepository) FindByIDOrPublicID(ctx context.Context, idOrPublicID string) (*models.User, error) {
	args := m.Called(ctx, idOrPublicID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// [VN] List giả lập truy vấn danh sách người dùng có hỗ trợ phân trang cursor
func (m *MockUserRepository) List(ctx context.Context, cursor string, limit int, search string, status string, countryCode string) ([]models.User, string, int64, error) {
	args := m.Called(ctx, cursor, limit, search, status, countryCode)
	var users []models.User
	if args.Get(0) != nil {
		users = args.Get(0).([]models.User)
	}
	return users, args.String(1), args.Get(2).(int64), args.Error(3)
}

// [VN] Update giả lập cập nhật thông tin thực thể người dùng
func (m *MockUserRepository) Update(ctx context.Context, tx *gorm.DB, user *models.User) error {
	args := m.Called(ctx, tx, user)
	return args.Error(0)
}

// [VN] Delete giả lập thao tác xóa mềm hoặc xóa cứng người dùng
func (m *MockUserRepository) Delete(ctx context.Context, tx *gorm.DB, id uuid.UUID, soft bool) error {
	args := m.Called(ctx, tx, id, soft)
	return args.Error(0)
}

// [VN] GetDB trả về con trỏ gorm.DB giả lập được thiết lập sẵn
func (m *MockUserRepository) GetDB() *gorm.DB {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*gorm.DB)
}

// [VN] MockCredentialRepository giả lập interface CredentialRepository
type MockCredentialRepository struct {
	mock.Mock
}

// [VN] Create giả lập tạo mới bản ghi thông tin xác thực mật khẩu
func (m *MockCredentialRepository) Create(ctx context.Context, tx *gorm.DB, creds *models.UserCredentials) error {
	args := m.Called(ctx, tx, creds)
	return args.Error(0)
}

// [VN] FindByUserID giả lập tìm kiếm thông tin xác thực theo UserID
func (m *MockCredentialRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.UserCredentials, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserCredentials), args.Error(1)
}

// [VN] Update giả lập cập nhật bản ghi xác thực
func (m *MockCredentialRepository) Update(ctx context.Context, tx *gorm.DB, creds *models.UserCredentials) error {
	args := m.Called(ctx, tx, creds)
	return args.Error(0)
}

// [VN] RecordFailedLogin giả lập ghi nhận số lần đăng nhập sai và thời gian khóa tài khoản
func (m *MockCredentialRepository) RecordFailedLogin(ctx context.Context, tx *gorm.DB, userID uuid.UUID, maxAttempts int, lockDuration time.Duration) (*models.UserCredentials, error) {
	args := m.Called(ctx, tx, userID, maxAttempts, lockDuration)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserCredentials), args.Error(1)
}

// [VN] ResetFailedLogin giả lập thiết lập lại trạng thái đăng nhập sai sau khi đăng nhập thành công
func (m *MockCredentialRepository) ResetFailedLogin(ctx context.Context, tx *gorm.DB, userID uuid.UUID) error {
	args := m.Called(ctx, tx, userID)
	return args.Error(0)
}

// [VN] IncrementTokenVersion giả lập tăng phiên bản token nhằm thu hồi toàn bộ các phiên làm việc
func (m *MockCredentialRepository) IncrementTokenVersion(ctx context.Context, tx *gorm.DB, userID uuid.UUID) error {
	args := m.Called(ctx, tx, userID)
	return args.Error(0)
}

// [VN] MockProfileRepository giả lập interface ProfileRepository
type MockProfileRepository struct {
	mock.Mock
}

// [VN] Create giả lập tạo hồ sơ cá nhân mới
func (m *MockProfileRepository) Create(ctx context.Context, tx *gorm.DB, profile *models.UserProfile) error {
	args := m.Called(ctx, tx, profile)
	return args.Error(0)
}

// [VN] FindByUserID giả lập tìm hồ sơ cá nhân theo UserID
func (m *MockProfileRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.UserProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserProfile), args.Error(1)
}

// [VN] Update giả lập cập nhật hồ sơ cá nhân
func (m *MockProfileRepository) Update(ctx context.Context, tx *gorm.DB, profile *models.UserProfile) error {
	args := m.Called(ctx, tx, profile)
	return args.Error(0)
}

// [VN] Upsert giả lập tạo mới hoặc cập nhật hồ sơ cá nhân
func (m *MockProfileRepository) Upsert(ctx context.Context, tx *gorm.DB, profile *models.UserProfile) error {
	args := m.Called(ctx, tx, profile)
	return args.Error(0)
}

// [VN] MockCompanyRepository giả lập interface CompanyRepository phục vụ kiểm thử đơn vị
type MockCompanyRepository struct {
	mock.Mock
}

// [VN] Create giả lập thao tác lưu công ty mới vào CSDL
func (m *MockCompanyRepository) Create(ctx context.Context, tx *gorm.DB, company *models.Company) error {
	args := m.Called(ctx, tx, company)
	return args.Error(0)
}

// [VN] FindByID giả lập tìm kiếm công ty theo UUID
func (m *MockCompanyRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Company, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Company), args.Error(1)
}

// [VN] FindByPublicID giả lập tìm kiếm công ty theo public_id
func (m *MockCompanyRepository) FindByPublicID(ctx context.Context, publicID string) (*models.Company, error) {
	args := m.Called(ctx, publicID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Company), args.Error(1)
}

// [VN] FindByCode giả lập tìm kiếm công ty theo mã code
func (m *MockCompanyRepository) FindByCode(ctx context.Context, code string) (*models.Company, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Company), args.Error(1)
}

// [VN] FindByTaxID giả lập tìm kiếm công ty theo mã số thuế
func (m *MockCompanyRepository) FindByTaxID(ctx context.Context, taxID string) (*models.Company, error) {
	args := m.Called(ctx, taxID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Company), args.Error(1)
}

// [VN] FindByIDOrPublicIDOrCode giả lập tra cứu công ty linh hoạt theo UUID, PublicID hoặc Code
func (m *MockCompanyRepository) FindByIDOrPublicIDOrCode(ctx context.Context, identifier string) (*models.Company, error) {
	args := m.Called(ctx, identifier)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Company), args.Error(1)
}

// [VN] List giả lập truy vấn danh sách công ty hỗ trợ phân trang cursor
func (m *MockCompanyRepository) List(
	ctx context.Context,
	cursor string,
	limit int,
	search string,
	status string,
	ownerID *uuid.UUID,
	countryCode string,
) ([]models.Company, string, int64, error) {
	args := m.Called(ctx, cursor, limit, search, status, ownerID, countryCode)
	var companies []models.Company
	if args.Get(0) != nil {
		companies = args.Get(0).([]models.Company)
	}
	return companies, args.String(1), args.Get(2).(int64), args.Error(3)
}

// [VN] Update giả lập cập nhật thông tin công ty
func (m *MockCompanyRepository) Update(ctx context.Context, tx *gorm.DB, company *models.Company) error {
	args := m.Called(ctx, tx, company)
	return args.Error(0)
}

// [VN] Delete giả lập xóa mềm hoặc xóa cứng công ty
func (m *MockCompanyRepository) Delete(ctx context.Context, tx *gorm.DB, id uuid.UUID, soft bool) error {
	args := m.Called(ctx, tx, id, soft)
	return args.Error(0)
}

// [VN] GetDB trả về con trỏ gorm.DB giả lập
func (m *MockCompanyRepository) GetDB() *gorm.DB {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*gorm.DB)
}

// [VN] newMockGormDB khởi tạo đối tượng gorm.DB giả lập thông qua sqlmock phục vụ kiểm thử Transaction
func newMockGormDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "failed to initialize sqlmock")

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err, "failed to initialize gorm with sqlmock")

	return gormDB, mock
}
