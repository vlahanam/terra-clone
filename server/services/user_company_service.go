package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vlahanam/terra-clone/common"
	"github.com/vlahanam/terra-clone/dto"
	"github.com/vlahanam/terra-clone/models"
	"github.com/vlahanam/terra-clone/repositories"
	"gorm.io/gorm"
)

// [VN] UserCompanyService định nghĩa các phương thức xử lý nghiệp vụ liên kết người dùng và công ty
type UserCompanyService interface {
	CreateUserCompany(ctx context.Context, req *dto.CreateUserCompanyRequest) (*dto.UserCompanyResponse, error)
	GetUserCompanyByID(ctx context.Context, identifier string) (*dto.UserCompanyResponse, error)
	ListUserCompanies(ctx context.Context, query *dto.ListUserCompaniesQuery) ([]dto.UserCompanyResponse, *common.Paging, error)
	UpdateUserCompany(ctx context.Context, identifier string, req *dto.UpdateUserCompanyRequest) (*dto.UserCompanyResponse, error)
	DeleteUserCompany(ctx context.Context, identifier string, hard bool) error
}

type userCompanyService struct {
	userCompanyRepo repositories.UserCompanyRepository
	userRepo        repositories.UserRepository
	companyRepo     repositories.CompanyRepository
}

// [VN] NewUserCompanyService khởi tạo UserCompanyService với các repository phụ thuộc
func NewUserCompanyService(
	userCompanyRepo repositories.UserCompanyRepository,
	userRepo repositories.UserRepository,
	companyRepo repositories.CompanyRepository,
) UserCompanyService {
	return &userCompanyService{
		userCompanyRepo: userCompanyRepo,
		userRepo:        userRepo,
		companyRepo:     companyRepo,
	}
}

// [VN] isValidRole kiểm tra xem vai trò được cung cấp có hợp lệ trong hệ thống hay không
func isValidRole(role string) bool {
	switch role {
	case models.UserCompanyRoleOwner,
		models.UserCompanyRoleAdmin,
		models.UserCompanyRoleMember,
		models.UserCompanyRoleGuest:
		return true
	default:
		return false
	}
}

// [VN] isValidStatus kiểm tra xem trạng thái liên kết có hợp lệ hay không
func isValidStatus(status string) bool {
	switch status {
	case models.UserCompanyStatusActive,
		models.UserCompanyStatusInactive,
		models.UserCompanyStatusPending,
		models.UserCompanyStatusSuspended,
		models.UserCompanyStatusDeleted:
		return true
	default:
		return false
	}
}

// [VN] CreateUserCompany liên kết một người dùng vào công ty sau khi thẩm định người dùng và công ty tồn tại
func (s *userCompanyService) CreateUserCompany(
	ctx context.Context,
	req *dto.CreateUserCompanyRequest,
) (*dto.UserCompanyResponse, error) {
	cleanUserID := strings.TrimSpace(req.UserID)
	if cleanUserID == "" {
		return nil, common.ErrBadRequest.WithReason("User ID cannot be empty")
	}

	cleanCompanyID := strings.TrimSpace(req.CompanyID)
	if cleanCompanyID == "" {
		return nil, common.ErrBadRequest.WithReason("Company ID cannot be empty")
	}

	// [VN] Tìm kiếm người dùng và kiểm tra tính hợp lệ
	user, err := s.userRepo.FindByIDOrPublicID(ctx, cleanUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound.
				WithReason("User not found in the system").
				WithDetail("user_id", cleanUserID)
		}
		return nil, common.ErrInternalServerError.
			WithReason("Failed to query user").
			WithTrace(err)
	}
	if user == nil {
		return nil, common.ErrNotFound.
			WithReason("User not found in the system").
			WithDetail("user_id", cleanUserID)
	}
	if user.Status == models.UserStatusDeleted {
		return nil, common.ErrBadRequest.
			WithReason("Cannot associate a deleted user to a company").
			WithDetail("user_status", user.Status)
	}

	// [VN] Tìm kiếm công ty theo ID, public_id hoặc mã code
	company, err := s.companyRepo.FindByIDOrPublicIDOrCode(ctx, cleanCompanyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound.
				WithReason("Company not found in the system").
				WithDetail("company_id", cleanCompanyID)
		}
		return nil, common.ErrInternalServerError.
			WithReason("Failed to query company").
			WithTrace(err)
	}
	if company == nil {
		return nil, common.ErrNotFound.
			WithReason("Company not found in the system").
			WithDetail("company_id", cleanCompanyID)
	}
	if company.Status == models.CompanyStatusDeleted {
		return nil, common.ErrBadRequest.
			WithReason("Cannot associate a user to a deleted company").
			WithDetail("company_status", company.Status)
	}

	// [VN] Chuẩn hóa và xác thực vai trò
	role := strings.ToLower(strings.TrimSpace(req.Role))
	if role == "" {
		role = models.UserCompanyRoleMember
	} else if !isValidRole(role) {
		return nil, common.ErrBadRequest.
			WithReason("Invalid role specified for user company association").
			WithDetail("role", req.Role)
	}

	// [VN] Chuẩn hóa và xác thực trạng thái
	status := strings.ToLower(strings.TrimSpace(req.Status))
	if status == "" {
		status = models.UserCompanyStatusActive
	} else if !isValidStatus(status) {
		return nil, common.ErrBadRequest.
			WithReason("Invalid status specified for user company association").
			WithDetail("status", req.Status)
	}

	// [VN] Kiểm tra xem liên kết giữa người dùng và công ty đã tồn tại hay chưa
	existing, err := s.userCompanyRepo.FindByUserAndCompany(ctx, user.ID, company.ID)
	if err == nil && existing != nil {
		// [VN] Nếu bản ghi trước đó đã bị xóa mềm, cho phép kích hoạt lại với thông tin mới
		if existing.Status == models.UserCompanyStatusDeleted {
			existing.Role = role
			existing.Status = status
			existing.JobTitle = req.JobTitle
			existing.Department = req.Department
			if req.Permissions != nil {
				existing.Permissions = models.JSONMap(req.Permissions)
			}
			if req.Metadata != nil {
				existing.Metadata = models.JSONMap(req.Metadata)
			}
			now := time.Now()
			existing.JoinedAt = &now
			existing.UpdatedAt = now

			if updateErr := s.userCompanyRepo.Update(ctx, nil, existing); updateErr != nil {
				return nil, common.ErrInternalServerError.
					WithReason("Failed to reactivate user company association").
					WithTrace(updateErr)
			}

			existing.User = user
			existing.Company = company
			return dto.ToUserCompanyResponse(existing), nil
		}

		return nil, common.ErrConflict.
			WithReason("User is already associated with this company").
			WithDetail("user_public_id", user.PublicID).
			WithDetail("company_public_id", company.PublicID)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.ErrInternalServerError.
			WithReason("Failed to verify existing user company association").
			WithTrace(err)
	}

	permissions := models.JSONMap(req.Permissions)
	if permissions == nil {
		permissions = make(models.JSONMap)
	}

	metadata := models.JSONMap(req.Metadata)
	if metadata == nil {
		metadata = make(models.JSONMap)
	}

	now := time.Now()
	newUC := &models.UserCompany{
		UserID:      user.ID,
		CompanyID:   company.ID,
		Role:        role,
		JobTitle:    req.JobTitle,
		Department:  req.Department,
		Status:      status,
		Permissions: permissions,
		Metadata:    metadata,
		JoinedAt:    &now,
	}

	if err := s.userCompanyRepo.Create(ctx, nil, newUC); err != nil {
		return nil, common.ErrInternalServerError.
			WithReason("Failed to create user company association").
			WithTrace(err)
	}

	newUC.User = user
	newUC.Company = company
	return dto.ToUserCompanyResponse(newUC), nil
}

// [VN] GetUserCompanyByID tìm kiếm thông tin liên kết UserCompany theo ID hoặc public_id
func (s *userCompanyService) GetUserCompanyByID(
	ctx context.Context,
	identifier string,
) (*dto.UserCompanyResponse, error) {
	cleanID := strings.TrimSpace(identifier)
	if cleanID == "" {
		return nil, common.ErrBadRequest.WithReason("User company identifier cannot be empty")
	}

	uc, err := s.userCompanyRepo.FindByIDOrPublicID(ctx, cleanID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound.
				WithReason("User company association not found").
				WithDetail("identifier", cleanID)
		}
		return nil, common.ErrInternalServerError.
			WithReason("Failed to query user company association").
			WithTrace(err)
	}
	if uc == nil {
		return nil, common.ErrNotFound.
			WithReason("User company association not found").
			WithDetail("identifier", cleanID)
	}

	return dto.ToUserCompanyResponse(uc), nil
}

// [VN] ListUserCompanies lấy danh sách liên kết hỗ trợ phân trang Cursor và bộ lọc tìm kiếm
func (s *userCompanyService) ListUserCompanies(
	ctx context.Context,
	query *dto.ListUserCompaniesQuery,
) ([]dto.UserCompanyResponse, *common.Paging, error) {
	var userID *uuid.UUID
	var companyID *uuid.UUID

	// [VN] Phân giải user_id từ query (chấp nhận UUID hoặc public_id)
	if cleanUser := strings.TrimSpace(query.UserID); cleanUser != "" {
		u, err := s.userRepo.FindByIDOrPublicID(ctx, cleanUser)
		if err == nil && u != nil {
			userID = &u.ID
		} else {
			// [VN] Nếu người dùng tìm kiếm không tồn tại, trả về danh sách rỗng
			return []dto.UserCompanyResponse{}, &common.Paging{
				Total:      0,
				FakeCursor: query.Cursor,
				NextCursor: "",
			}, nil
		}
	}

	// [VN] Phân giải company_id từ query (chấp nhận UUID, public_id hoặc code)
	if cleanComp := strings.TrimSpace(query.CompanyID); cleanComp != "" {
		c, err := s.companyRepo.FindByIDOrPublicIDOrCode(ctx, cleanComp)
		if err == nil && c != nil {
			companyID = &c.ID
		} else {
			// [VN] Nếu công ty tìm kiếm không tồn tại, trả về danh sách rỗng
			return []dto.UserCompanyResponse{}, &common.Paging{
				Total:      0,
				FakeCursor: query.Cursor,
				NextCursor: "",
			}, nil
		}
	}

	limit := query.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	results, nextCursor, total, err := s.userCompanyRepo.List(
		ctx,
		query.Cursor,
		limit,
		userID,
		companyID,
		query.Role,
		query.Status,
		query.Search,
	)
	if err != nil {
		return nil, nil, common.ErrInternalServerError.
			WithReason("Failed to retrieve user company list").
			WithTrace(err)
	}

	responses := make([]dto.UserCompanyResponse, 0, len(results))
	for i := range results {
		res := dto.ToUserCompanyResponse(&results[i])
		if res != nil {
			responses = append(responses, *res)
		}
	}

	paging := &common.Paging{
		Total:      total,
		FakeCursor: query.Cursor,
		NextCursor: nextCursor,
	}

	return responses, paging, nil
}

// [VN] UpdateUserCompany cập nhật thông tin vai trò, chức danh, phòng ban và cấu hình liên kết
func (s *userCompanyService) UpdateUserCompany(
	ctx context.Context,
	identifier string,
	req *dto.UpdateUserCompanyRequest,
) (*dto.UserCompanyResponse, error) {
	cleanID := strings.TrimSpace(identifier)
	if cleanID == "" {
		return nil, common.ErrBadRequest.WithReason("User company identifier cannot be empty")
	}

	uc, err := s.userCompanyRepo.FindByIDOrPublicID(ctx, cleanID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound.
				WithReason("User company association not found").
				WithDetail("identifier", cleanID)
		}
		return nil, common.ErrInternalServerError.
			WithReason("Failed to query user company association").
			WithTrace(err)
	}
	if uc == nil {
		return nil, common.ErrNotFound.
			WithReason("User company association not found").
			WithDetail("identifier", cleanID)
	}

	// [VN] Cập nhật vai trò nếu được cung cấp
	if req.Role != nil {
		newRole := strings.ToLower(strings.TrimSpace(*req.Role))
		if !isValidRole(newRole) {
			return nil, common.ErrBadRequest.
				WithReason("Invalid role specified for user company association").
				WithDetail("role", *req.Role)
		}
		uc.Role = newRole
	}

	// [VN] Cập nhật chức danh công việc nếu được cung cấp
	if req.JobTitle != nil {
		uc.JobTitle = req.JobTitle
	}

	// [VN] Cập nhật phòng ban nếu được cung cấp
	if req.Department != nil {
		uc.Department = req.Department
	}

	// [VN] Cập nhật trạng thái nếu được cung cấp
	if req.Status != nil {
		newStatus := strings.ToLower(strings.TrimSpace(*req.Status))
		if !isValidStatus(newStatus) {
			return nil, common.ErrBadRequest.
				WithReason("Invalid status specified for user company association").
				WithDetail("status", *req.Status)
		}
		uc.Status = newStatus
	}

	// [VN] Cập nhật quyền hạn JSONB nếu có
	if req.Permissions != nil {
		if uc.Permissions == nil {
			uc.Permissions = make(models.JSONMap)
		}
		for k, v := range req.Permissions {
			uc.Permissions[k] = v
		}
	}

	// [VN] Cập nhật cấu hình mở rộng metadata JSONB nếu có
	if req.Metadata != nil {
		if uc.Metadata == nil {
			uc.Metadata = make(models.JSONMap)
		}
		for k, v := range req.Metadata {
			uc.Metadata[k] = v
		}
	}

	uc.UpdatedAt = time.Now()

	if err := s.userCompanyRepo.Update(ctx, nil, uc); err != nil {
		return nil, common.ErrInternalServerError.
			WithReason("Failed to update user company association").
			WithTrace(err)
	}

	return dto.ToUserCompanyResponse(uc), nil
}

// [VN] DeleteUserCompany thực hiện soft-delete hoặc xóa cứng liên kết giữa người dùng và công ty
func (s *userCompanyService) DeleteUserCompany(
	ctx context.Context,
	identifier string,
	hard bool,
) error {
	cleanID := strings.TrimSpace(identifier)
	if cleanID == "" {
		return common.ErrBadRequest.WithReason("User company identifier cannot be empty")
	}

	uc, err := s.userCompanyRepo.FindByIDOrPublicID(ctx, cleanID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrNotFound.
				WithReason("User company association not found").
				WithDetail("identifier", cleanID)
		}
		return common.ErrInternalServerError.
			WithReason("Failed to query user company association").
			WithTrace(err)
	}
	if uc == nil {
		return common.ErrNotFound.
			WithReason("User company association not found").
			WithDetail("identifier", cleanID)
	}

	if err := s.userCompanyRepo.Delete(ctx, nil, uc.ID, !hard); err != nil {
		return common.ErrInternalServerError.
			WithReason("Failed to delete user company association").
			WithTrace(err)
	}

	return nil
}
