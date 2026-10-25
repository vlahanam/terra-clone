package repositories

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/vlahanam/terra-clone/models"
	"gorm.io/gorm"
)

// [VN] UserCompanyRepository định nghĩa interface thao tác dữ liệu với bảng user_companies
type UserCompanyRepository interface {
	Create(ctx context.Context, tx *gorm.DB, uc *models.UserCompany) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.UserCompany, error)
	FindByPublicID(ctx context.Context, publicID string) (*models.UserCompany, error)
	FindByIDOrPublicID(ctx context.Context, identifier string) (*models.UserCompany, error)
	FindByUserAndCompany(ctx context.Context, userID uuid.UUID, companyID uuid.UUID) (*models.UserCompany, error)
	List(ctx context.Context, cursor string, limit int, userID *uuid.UUID, companyID *uuid.UUID, role string, status string, search string) ([]models.UserCompany, string, int64, error)
	Update(ctx context.Context, tx *gorm.DB, uc *models.UserCompany) error
	Delete(ctx context.Context, tx *gorm.DB, id uuid.UUID, soft bool) error
	GetDB() *gorm.DB
}

type userCompanyRepository struct {
	db *gorm.DB
}

// [VN] NewUserCompanyRepository khởi tạo UserCompanyRepository với kết nối gorm.DB
func NewUserCompanyRepository(db *gorm.DB) UserCompanyRepository {
	return &userCompanyRepository{db: db}
}

// [VN] GetDB trả về con trỏ gorm.DB cơ sở
func (r *userCompanyRepository) GetDB() *gorm.DB {
	return r.db
}

// [VN] getDB hỗ trợ tái sử dụng database transaction nếu tx được truyền vào
func (r *userCompanyRepository) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

// [VN] Create tạo mới bản ghi liên kết user_companies trong cơ sở dữ liệu
func (r *userCompanyRepository) Create(ctx context.Context, tx *gorm.DB, uc *models.UserCompany) error {
	return r.getDB(tx).WithContext(ctx).Create(uc).Error
}

// [VN] FindByID tìm kiếm liên kết theo khóa chính UUID kèm nạp thông tin User và Company
func (r *userCompanyRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.UserCompany, error) {
	var uc models.UserCompany
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Company").
		Where("id = ?", id).
		First(&uc).Error
	if err != nil {
		return nil, err
	}
	return &uc, nil
}

// [VN] FindByPublicID tìm kiếm liên kết theo mã định danh công khai (public_id)
func (r *userCompanyRepository) FindByPublicID(ctx context.Context, publicID string) (*models.UserCompany, error) {
	var uc models.UserCompany
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Company").
		Where("public_id = ?", publicID).
		First(&uc).Error
	if err != nil {
		return nil, err
	}
	return &uc, nil
}

// [VN] FindByIDOrPublicID tự động phát hiện tham số là UUID hay public_id để tìm kiếm
func (r *userCompanyRepository) FindByIDOrPublicID(ctx context.Context, identifier string) (*models.UserCompany, error) {
	cleanIdentifier := strings.TrimSpace(identifier)
	if parsedUUID, err := uuid.Parse(cleanIdentifier); err == nil {
		return r.FindByID(ctx, parsedUUID)
	}
	return r.FindByPublicID(ctx, cleanIdentifier)
}

// [VN] FindByUserAndCompany tìm liên kết giữa một người dùng cụ thể và một công ty cụ thể
func (r *userCompanyRepository) FindByUserAndCompany(ctx context.Context, userID uuid.UUID, companyID uuid.UUID) (*models.UserCompany, error) {
	var uc models.UserCompany
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Company").
		Where("user_id = ? AND company_id = ?", userID, companyID).
		First(&uc).Error
	if err != nil {
		return nil, err
	}
	return &uc, nil
}

// [VN] List lấy danh sách liên kết hỗ trợ phân trang Cursor, lọc theo user, company, role, status và tìm kiếm
func (r *userCompanyRepository) List(
	ctx context.Context,
	cursor string,
	limit int,
	userID *uuid.UUID,
	companyID *uuid.UUID,
	role string,
	status string,
	search string,
) ([]models.UserCompany, string, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := r.db.WithContext(ctx).Model(&models.UserCompany{})

	if status != "" {
		query = query.Where("user_companies.status = ?", status)
	} else {
		// [VN] Mặc định loại bỏ các bản ghi đã bị xóa mềm khỏi danh sách
		query = query.Where("user_companies.status != ?", models.UserCompanyStatusDeleted)
	}

	if userID != nil {
		query = query.Where("user_companies.user_id = ?", *userID)
	}

	if companyID != nil {
		query = query.Where("user_companies.company_id = ?", *companyID)
	}

	if role != "" {
		query = query.Where("user_companies.role = ?", strings.ToLower(strings.TrimSpace(role)))
	}

	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		query = query.Joins("LEFT JOIN users ON users.id = user_companies.user_id").
			Joins("LEFT JOIN companies ON companies.id = user_companies.company_id").
			Where(
				"(LOWER(user_companies.job_title) LIKE ? OR LOWER(user_companies.department) LIKE ? OR LOWER(user_companies.role) LIKE ? OR user_companies.public_id = ? OR LOWER(users.email) LIKE ? OR LOWER(users.full_name) LIKE ? OR LOWER(companies.name) LIKE ? OR LOWER(companies.code) LIKE ?)",
				searchPattern, searchPattern, searchPattern, search, searchPattern, searchPattern, searchPattern, searchPattern,
			)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, "", 0, err
	}

	dataQuery := query.Preload("User").Preload("Company").Order("user_companies.id DESC").Limit(limit)

	if cursor != "" {
		if cursorUUID, err := uuid.Parse(cursor); err == nil {
			dataQuery = dataQuery.Where("user_companies.id < ?", cursorUUID)
		} else {
			// [VN] Tra cứu khóa chính UUID tương ứng từ public_id để làm mốc phân trang con trỏ
			var cursorUC models.UserCompany
			if err := r.db.WithContext(ctx).Select("id").Where("public_id = ?", cursor).First(&cursorUC).Error; err == nil {
				dataQuery = dataQuery.Where("user_companies.id < ?", cursorUC.ID)
			}
		}
	}

	var results []models.UserCompany
	if err := dataQuery.Find(&results).Error; err != nil {
		return nil, "", 0, err
	}

	var nextCursor string
	if len(results) == limit {
		nextCursor = results[len(results)-1].PublicID
	}

	return results, nextCursor, total, nil
}

// [VN] Update cập nhật thông tin liên kết UserCompany
func (r *userCompanyRepository) Update(ctx context.Context, tx *gorm.DB, uc *models.UserCompany) error {
	return r.getDB(tx).WithContext(ctx).Save(uc).Error
}

// [VN] Delete thực hiện soft-delete hoặc xóa cứng liên kết UserCompany
func (r *userCompanyRepository) Delete(ctx context.Context, tx *gorm.DB, id uuid.UUID, soft bool) error {
	if soft {
		return r.getDB(tx).WithContext(ctx).Model(&models.UserCompany{}).
			Where("id = ?", id).
			Update("status", models.UserCompanyStatusDeleted).Error
	}
	return r.getDB(tx).WithContext(ctx).Where("id = ?", id).Delete(&models.UserCompany{}).Error
}
