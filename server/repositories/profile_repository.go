package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/vlahanam/terra-clone/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// [VN] ProfileRepository định nghĩa interface thao tác dữ liệu với bảng user_profiles
type ProfileRepository interface {
	Create(ctx context.Context, tx *gorm.DB, profile *models.UserProfile) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.UserProfile, error)
	Update(ctx context.Context, tx *gorm.DB, profile *models.UserProfile) error
	Upsert(ctx context.Context, tx *gorm.DB, profile *models.UserProfile) error
}

type profileRepository struct {
	db *gorm.DB
}

// [VN] NewProfileRepository khởi tạo ProfileRepository
func NewProfileRepository(db *gorm.DB) ProfileRepository {
	return &profileRepository{db: db}
}

// [VN] getDB hỗ trợ tái sử dụng database transaction nếu tx được truyền vào
func (r *profileRepository) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

// [VN] Create tạo mới thông tin hồ sơ người dùng
func (r *profileRepository) Create(ctx context.Context, tx *gorm.DB, profile *models.UserProfile) error {
	return r.getDB(tx).WithContext(ctx).Create(profile).Error
}

// [VN] FindByUserID tìm kiếm hồ sơ người dùng theo UserID
func (r *profileRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.UserProfile, error) {
	var profile models.UserProfile
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// [VN] Update cập nhật thông tin hồ sơ người dùng
func (r *profileRepository) Update(ctx context.Context, tx *gorm.DB, profile *models.UserProfile) error {
	return r.getDB(tx).WithContext(ctx).Save(profile).Error
}

// [VN] Upsert thêm mới hoặc cập nhật hồ sơ người dùng với mệnh đề ON CONFLICT DO UPDATE
func (r *profileRepository) Upsert(ctx context.Context, tx *gorm.DB, profile *models.UserProfile) error {
	return r.getDB(tx).WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"date_of_birth", "gender", "phone_number", "address", "preferences", "updated_at"}),
	}).Create(profile).Error
}
