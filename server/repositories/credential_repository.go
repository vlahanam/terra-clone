package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vlahanam/terra-clone/models"
	"gorm.io/gorm"
)

// [VN] CredentialRepository định nghĩa interface thao tác dữ liệu với bảng user_credentials
type CredentialRepository interface {
	Create(ctx context.Context, tx *gorm.DB, creds *models.UserCredentials) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.UserCredentials, error)
	Update(ctx context.Context, tx *gorm.DB, creds *models.UserCredentials) error
	RecordFailedLogin(ctx context.Context, tx *gorm.DB, userID uuid.UUID, maxAttempts int, lockDuration time.Duration) (*models.UserCredentials, error)
	ResetFailedLogin(ctx context.Context, tx *gorm.DB, userID uuid.UUID) error
	IncrementTokenVersion(ctx context.Context, tx *gorm.DB, userID uuid.UUID) error
}

type credentialRepository struct {
	db *gorm.DB
}

// [VN] NewCredentialRepository khởi tạo thể hiện triển khai của CredentialRepository
func NewCredentialRepository(db *gorm.DB) CredentialRepository {
	return &credentialRepository{db: db}
}

// [VN] getDB hỗ trợ tái sử dụng database transaction nếu có truyền vào tx
func (r *credentialRepository) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

// [VN] Create thêm bản ghi thông tin xác thực mới
func (r *credentialRepository) Create(ctx context.Context, tx *gorm.DB, creds *models.UserCredentials) error {
	return r.getDB(tx).WithContext(ctx).Create(creds).Error
}

// [VN] FindByUserID tìm kiếm thông tin xác thực theo UserID
func (r *credentialRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.UserCredentials, error) {
	var creds models.UserCredentials
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&creds).Error
	if err != nil {
		return nil, err
	}
	return &creds, nil
}

// [VN] Update cập nhật toàn bộ thông tin xác thực
func (r *credentialRepository) Update(ctx context.Context, tx *gorm.DB, creds *models.UserCredentials) error {
	return r.getDB(tx).WithContext(ctx).Save(creds).Error
}

// [VN] RecordFailedLogin ghi nhận lần đăng nhập sai và tự động tạm khóa nếu vượt quá số lần cho phép
func (r *credentialRepository) RecordFailedLogin(ctx context.Context, tx *gorm.DB, userID uuid.UUID, maxAttempts int, lockDuration time.Duration) (*models.UserCredentials, error) {
	creds, err := r.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	creds.FailedLoginAttempts++
	if creds.FailedLoginAttempts >= maxAttempts {
		lockedTime := time.Now().Add(lockDuration)
		creds.LockedUntil = &lockedTime
	}

	if err := r.getDB(tx).WithContext(ctx).Save(creds).Error; err != nil {
		return nil, err
	}

	return creds, nil
}

// [VN] ResetFailedLogin xóa số lần đăng nhập sai và cập nhật thời điểm đăng nhập thành công
func (r *credentialRepository) ResetFailedLogin(ctx context.Context, tx *gorm.DB, userID uuid.UUID) error {
	now := time.Now()
	return r.getDB(tx).WithContext(ctx).Model(&models.UserCredentials{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"failed_login_attempts": 0,
			"locked_until":          nil,
			"last_login_at":         now,
		}).Error
}

// [VN] IncrementTokenVersion tăng token_version để vô hiệu hóa tất cả token cũ
func (r *credentialRepository) IncrementTokenVersion(ctx context.Context, tx *gorm.DB, userID uuid.UUID) error {
	return r.getDB(tx).WithContext(ctx).Model(&models.UserCredentials{}).
		Where("user_id = ?", userID).
		Update("token_version", gorm.Expr("token_version + 1")).Error
}
