package dto

import (
	"time"

	"github.com/vlahanam/terra-clone/models"
)

// [VN] CreateUserCompanyRequest định nghĩa dữ liệu đầu vào khi liên kết người dùng với công ty
type CreateUserCompanyRequest struct {
	UserID      string                 `json:"user_id" binding:"required"`
	CompanyID   string                 `json:"company_id" binding:"required"`
	Role        string                 `json:"role,omitempty"`
	JobTitle    *string                `json:"job_title,omitempty"`
	Department  *string                `json:"department,omitempty"`
	Status      string                 `json:"status,omitempty"`
	Permissions map[string]interface{} `json:"permissions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// [VN] UpdateUserCompanyRequest định nghĩa dữ liệu cập nhật liên kết giữa người dùng và công ty
type UpdateUserCompanyRequest struct {
	Role        *string                `json:"role,omitempty"`
	JobTitle    *string                `json:"job_title,omitempty"`
	Department  *string                `json:"department,omitempty"`
	Status      *string                `json:"status,omitempty"`
	Permissions map[string]interface{} `json:"permissions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// [VN] ListUserCompaniesQuery định nghĩa các tham số lọc, tìm kiếm và phân trang cursor cho danh sách liên kết
type ListUserCompaniesQuery struct {
	Cursor    string `form:"cursor"`
	Limit     int    `form:"limit"`
	UserID    string `form:"user_id"`
	CompanyID string `form:"company_id"`
	Role      string `form:"role"`
	Status    string `form:"status"`
	Search    string `form:"search"`
}

// [VN] UserCompanyUserResponse định nghĩa thông tin tóm tắt của người dùng trong liên kết (trường ID mang giá trị public_id để bảo mật)
type UserCompanyUserResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FullName  string  `json:"full_name"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	Status    string  `json:"status"`
}

// [VN] UserCompanyCompanyResponse định nghĩa thông tin tóm tắt của công ty trong liên kết (trường ID mang giá trị public_id để bảo mật)
type UserCompanyCompanyResponse struct {
	ID      string  `json:"id"`
	Code    string  `json:"code"`
	Name    string  `json:"name"`
	LogoURL *string `json:"logo_url,omitempty"`
	Status  string  `json:"status"`
}

// [VN] UserCompanyResponse định nghĩa cấu trúc chi tiết trả về của liên kết UserCompany (các trường ID mang giá trị public_id để bảo mật)
type UserCompanyResponse struct {
	ID          string                      `json:"id"`
	UserID      string                      `json:"user_id"`
	CompanyID   string                      `json:"company_id"`
	Role        string                      `json:"role"`
	JobTitle    *string                     `json:"job_title,omitempty"`
	Department  *string                     `json:"department,omitempty"`
	Status      string                      `json:"status"`
	Permissions map[string]interface{}      `json:"permissions"`
	Metadata    map[string]interface{}      `json:"metadata"`
	JoinedAt    *time.Time                  `json:"joined_at,omitempty"`
	CreatedAt   time.Time                   `json:"created_at"`
	UpdatedAt   time.Time                   `json:"updated_at"`
	User        *UserCompanyUserResponse    `json:"user,omitempty"`
	Company     *UserCompanyCompanyResponse `json:"company,omitempty"`
}

// [VN] ToUserCompanyResponse chuyển đổi entity models.UserCompany sang DTO UserCompanyResponse
func ToUserCompanyResponse(uc *models.UserCompany) *UserCompanyResponse {
	if uc == nil {
		return nil
	}

	permissions := map[string]interface{}(uc.Permissions)
	if permissions == nil {
		permissions = make(map[string]interface{})
	}

	metadata := map[string]interface{}(uc.Metadata)
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	res := &UserCompanyResponse{
		ID:          uc.PublicID,
		Role:        uc.Role,
		JobTitle:    uc.JobTitle,
		Department:  uc.Department,
		Status:      uc.Status,
		Permissions: permissions,
		Metadata:    metadata,
		JoinedAt:    uc.JoinedAt,
		CreatedAt:   uc.CreatedAt,
		UpdatedAt:   uc.UpdatedAt,
	}

	if uc.User != nil {
		res.User = &UserCompanyUserResponse{
			ID:        uc.User.PublicID,
			Email:     uc.User.Email,
			FullName:  uc.User.FullName,
			AvatarURL: uc.User.AvatarURL,
			Status:    uc.User.Status,
		}
		res.UserID = uc.User.PublicID
	}

	if uc.Company != nil {
		res.Company = &UserCompanyCompanyResponse{
			ID:      uc.Company.PublicID,
			Code:    uc.Company.Code,
			Name:    uc.Company.Name,
			LogoURL: uc.Company.LogoURL,
			Status:  uc.Company.Status,
		}
		res.CompanyID = uc.Company.PublicID
	}

	return res
}
