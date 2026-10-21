package repositories

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/vlahanam/terra-clone/models"
	"gorm.io/gorm"
)

// [VN] UserRepository định nghĩa interface thao tác dữ liệu với bảng users
type UserRepository interface {
	Create(ctx context.Context, tx *gorm.DB, user *models.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	FindByPublicID(ctx context.Context, publicID string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByIDOrPublicID(ctx context.Context, idOrPublicID string) (*models.User, error)
	List(ctx context.Context, cursor string, limit int, search string, status string) ([]models.User, string, int64, error)
	Update(ctx context.Context, tx *gorm.DB, user *models.User) error
	Delete(ctx context.Context, tx *gorm.DB, id uuid.UUID, soft bool) error
	GetDB() *gorm.DB
}

type userRepository struct {
	db *gorm.DB
}

// [VN] NewUserRepository khởi tạo UserRepository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// [VN] GetDB trả về con trỏ kết nối gorm.DB cơ sở
func (r *userRepository) GetDB() *gorm.DB {
	return r.db
}

// [VN] getDB hỗ trợ tái sử dụng database transaction nếu tx được truyền vào
func (r *userRepository) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

// [VN] Create tạo mới người dùng trong cơ sở dữ liệu
func (r *userRepository) Create(ctx context.Context, tx *gorm.DB, user *models.User) error {
	return r.getDB(tx).WithContext(ctx).Create(user).Error
}

// [VN] FindByID tìm kiếm người dùng theo khóa chính UUID
func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Preload("Profile").
		Preload("Credentials").
		Where("id = ?", id).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// [VN] FindByPublicID tìm kiếm người dùng theo mã định danh công khai
func (r *userRepository) FindByPublicID(ctx context.Context, publicID string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Preload("Profile").
		Preload("Credentials").
		Where("public_id = ?", publicID).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// [VN] FindByEmail tìm kiếm người dùng theo địa chỉ email không phân biệt hoa thường
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Preload("Profile").
		Preload("Credentials").
		Where("LOWER(email) = LOWER(?)", strings.TrimSpace(email)).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// [VN] FindByIDOrPublicID tự động phát hiện tham số là UUID hay public_id để tìm kiếm
func (r *userRepository) FindByIDOrPublicID(ctx context.Context, idOrPublicID string) (*models.User, error) {
	parsedUUID, err := uuid.Parse(idOrPublicID)
	if err == nil {
		return r.FindByID(ctx, parsedUUID)
	}
	return r.FindByPublicID(ctx, idOrPublicID)
}

// [VN] List lấy danh sách người dùng hỗ trợ phân trang Cursor và tìm kiếm đa trường
func (r *userRepository) List(ctx context.Context, cursor string, limit int, search string, status string) ([]models.User, string, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := r.db.WithContext(ctx).Model(&models.User{})

	if status != "" {
		query = query.Where("status = ?", status)
	} else {
		// [VN] Mặc định không hiển thị tài khoản đã xóa nếu không chỉ định rõ
		query = query.Where("status != ?", models.UserStatusDeleted)
	}

	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		query = query.Where("(LOWER(email) LIKE ? OR LOWER(full_name) LIKE ? OR public_id = ?)", searchPattern, searchPattern, search)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, "", 0, err
	}

	dataQuery := query.Preload("Profile").Order("id DESC").Limit(limit)

	if cursor != "" {
		if cursorUUID, err := uuid.Parse(cursor); err == nil {
			dataQuery = dataQuery.Where("id < ?", cursorUUID)
		}
	}

	var users []models.User
	if err := dataQuery.Find(&users).Error; err != nil {
		return nil, "", 0, err
	}

	var nextCursor string
	if len(users) == limit {
		nextCursor = users[len(users)-1].ID.String()
	}

	return users, nextCursor, total, nil
}

// [VN] Update cập nhật thông tin người dùng
func (r *userRepository) Update(ctx context.Context, tx *gorm.DB, user *models.User) error {
	return r.getDB(tx).WithContext(ctx).Save(user).Error
}

// [VN] Delete thực hiện soft-delete hoặc xóa cứng người dùng
func (r *userRepository) Delete(ctx context.Context, tx *gorm.DB, id uuid.UUID, soft bool) error {
	if soft {
		return r.getDB(tx).WithContext(ctx).Model(&models.User{}).
			Where("id = ?", id).
			Update("status", models.UserStatusDeleted).Error
	}
	return r.getDB(tx).WithContext(ctx).Where("id = ?", id).Delete(&models.User{}).Error
}
