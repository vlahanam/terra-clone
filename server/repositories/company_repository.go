package repositories

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/vlahanam/terra-clone/models"
	"gorm.io/gorm"
)

// [VN] CompanyRepository định nghĩa interface thao tác dữ liệu với bảng companies
type CompanyRepository interface {
	Create(ctx context.Context, tx *gorm.DB, company *models.Company) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Company, error)
	FindByPublicID(ctx context.Context, publicID string) (*models.Company, error)
	FindByCode(ctx context.Context, code string) (*models.Company, error)
	FindByTaxID(ctx context.Context, taxID string) (*models.Company, error)
	FindByIDOrPublicIDOrCode(ctx context.Context, identifier string) (*models.Company, error)
	List(ctx context.Context, cursor string, limit int, search string, status string, ownerID *uuid.UUID, countryCode string) ([]models.Company, string, int64, error)
	Update(ctx context.Context, tx *gorm.DB, company *models.Company) error
	Delete(ctx context.Context, tx *gorm.DB, id uuid.UUID, soft bool) error
	GetDB() *gorm.DB
}

type companyRepository struct {
	db *gorm.DB
}

// [VN] NewCompanyRepository khởi tạo CompanyRepository với kết nối gorm.DB
func NewCompanyRepository(db *gorm.DB) CompanyRepository {
	return &companyRepository{db: db}
}

// [VN] GetDB trả về con trỏ gorm.DB cơ sở
func (r *companyRepository) GetDB() *gorm.DB {
	return r.db
}

// [VN] getDB hỗ trợ tái sử dụng database transaction nếu tx được truyền vào
func (r *companyRepository) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

// [VN] Create tạo mới bản ghi công ty trong cơ sở dữ liệu
func (r *companyRepository) Create(ctx context.Context, tx *gorm.DB, company *models.Company) error {
	return r.getDB(tx).WithContext(ctx).Create(company).Error
}

// [VN] FindByID tìm kiếm công ty theo khóa chính UUID
func (r *companyRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Company, error) {
	var company models.Company
	err := r.db.WithContext(ctx).
		Preload("Owner").
		Where("id = ?", id).
		First(&company).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

// [VN] FindByPublicID tìm kiếm công ty theo mã định danh công khai (public_id)
func (r *companyRepository) FindByPublicID(ctx context.Context, publicID string) (*models.Company, error) {
	var company models.Company
	err := r.db.WithContext(ctx).
		Preload("Owner").
		Where("public_id = ?", publicID).
		First(&company).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

// [VN] FindByCode tìm kiếm công ty theo mã code duy nhất (không phân biệt chữ hoa thường)
func (r *companyRepository) FindByCode(ctx context.Context, code string) (*models.Company, error) {
	var company models.Company
	err := r.db.WithContext(ctx).
		Preload("Owner").
		Where("LOWER(code) = LOWER(?)", strings.TrimSpace(code)).
		First(&company).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

// [VN] FindByTaxID tìm kiếm công ty theo mã số thuế
func (r *companyRepository) FindByTaxID(ctx context.Context, taxID string) (*models.Company, error) {
	var company models.Company
	err := r.db.WithContext(ctx).
		Preload("Owner").
		Where("tax_id = ?", strings.TrimSpace(taxID)).
		First(&company).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

// [VN] FindByIDOrPublicIDOrCode tìm kiếm công ty linh hoạt theo UUID, PublicID hoặc Code
func (r *companyRepository) FindByIDOrPublicIDOrCode(ctx context.Context, identifier string) (*models.Company, error) {
	cleanIdentifier := strings.TrimSpace(identifier)

	// [VN] Thử phân tích cú pháp UUID nếu identifier hợp lệ chuẩn UUID
	if parsedUUID, err := uuid.Parse(cleanIdentifier); err == nil {
		return r.FindByID(ctx, parsedUUID)
	}

	// [VN] Nếu có tiền tố cmp_ thì ưu tiên tra cứu theo public_id
	if strings.HasPrefix(cleanIdentifier, "cmp_") {
		return r.FindByPublicID(ctx, cleanIdentifier)
	}

	// [VN] Kiểm tra qua code trước, nếu không tìm thấy thì thử theo public_id
	company, err := r.FindByCode(ctx, cleanIdentifier)
	if err == nil && company != nil {
		return company, nil
	}

	return r.FindByPublicID(ctx, cleanIdentifier)
}

// [VN] List lấy danh sách công ty hỗ trợ phân trang Cursor và tìm kiếm đa trường
func (r *companyRepository) List(
	ctx context.Context,
	cursor string,
	limit int,
	search string,
	status string,
	ownerID *uuid.UUID,
	countryCode string,
) ([]models.Company, string, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := r.db.WithContext(ctx).Model(&models.Company{})

	if status != "" {
		query = query.Where("status = ?", status)
	} else {
		// [VN] Mặc định loại bỏ các công ty đã bị xóa mềm khỏi danh sách
		query = query.Where("status != ?", models.CompanyStatusDeleted)
	}

	if ownerID != nil {
		query = query.Where("owner_id = ?", *ownerID)
	}

	if countryCode != "" {
		query = query.Where("country_code = ?", strings.ToUpper(strings.TrimSpace(countryCode)))
	}

	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		query = query.Where(
			"(LOWER(name) LIKE ? OR LOWER(code) LIKE ? OR LOWER(COALESCE(tax_id, '')) LIKE ? OR LOWER(COALESCE(email, '')) LIKE ? OR public_id = ?)",
			searchPattern, searchPattern, searchPattern, searchPattern, search,
		)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, "", 0, err
	}

	dataQuery := query.Preload("Owner").Order("id DESC").Limit(limit)

	if cursor != "" {
		if cursorUUID, err := uuid.Parse(cursor); err == nil {
			dataQuery = dataQuery.Where("id < ?", cursorUUID)
		}
	}

	var companies []models.Company
	if err := dataQuery.Find(&companies).Error; err != nil {
		return nil, "", 0, err
	}

	var nextCursor string
	if len(companies) == limit {
		nextCursor = companies[len(companies)-1].ID.String()
	}

	return companies, nextCursor, total, nil
}

// [VN] Update cập nhật toàn bộ thông tin công ty
func (r *companyRepository) Update(ctx context.Context, tx *gorm.DB, company *models.Company) error {
	return r.getDB(tx).WithContext(ctx).Save(company).Error
}

// [VN] Delete thực hiện soft-delete hoặc xóa cứng công ty
func (r *companyRepository) Delete(ctx context.Context, tx *gorm.DB, id uuid.UUID, soft bool) error {
	if soft {
		return r.getDB(tx).WithContext(ctx).Model(&models.Company{}).
			Where("id = ?", id).
			Update("status", models.CompanyStatusDeleted).Error
	}
	return r.getDB(tx).WithContext(ctx).Where("id = ?", id).Delete(&models.Company{}).Error
}
