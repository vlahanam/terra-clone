package dto

import (
	"time"

	"github.com/vlahanam/terra-clone/models"
)

// [VN] CreateCompanyRequest định nghĩa dữ liệu đầu vào khi tạo mới công ty
type CreateCompanyRequest struct {
	Name            string                 `json:"name" binding:"required"`
	Code            string                 `json:"code" binding:"required,min=2,max=50"`
	LegalName       *string                `json:"legal_name,omitempty"`
	TaxID           *string                `json:"tax_id,omitempty"`
	Email           *string                `json:"email,omitempty" binding:"omitempty,email"`
	Phone           *string                `json:"phone,omitempty"`
	Address         *string                `json:"address,omitempty"`
	City            *string                `json:"city,omitempty"`
	StateOrProvince *string                `json:"state_or_province,omitempty"`
	PostalCode      *string                `json:"postal_code,omitempty"`
	CountryCode     *string                `json:"country_code,omitempty"`
	Timezone        *string                `json:"timezone,omitempty"`
	Currency        *string                `json:"currency,omitempty"`
	Website         *string                `json:"website,omitempty"`
	LogoURL         *string                `json:"logo_url,omitempty"`
	Description     *string                `json:"description,omitempty"`
	Status          string                 `json:"status,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// [VN] UpdateCompanyRequest định nghĩa dữ liệu cập nhật thông tin công ty
type UpdateCompanyRequest struct {
	Name            *string                `json:"name,omitempty"`
	Code            *string                `json:"code,omitempty"`
	LegalName       *string                `json:"legal_name,omitempty"`
	TaxID           *string                `json:"tax_id,omitempty"`
	Email           *string                `json:"email,omitempty" binding:"omitempty,email"`
	Phone           *string                `json:"phone,omitempty"`
	Address         *string                `json:"address,omitempty"`
	City            *string                `json:"city,omitempty"`
	StateOrProvince *string                `json:"state_or_province,omitempty"`
	PostalCode      *string                `json:"postal_code,omitempty"`
	CountryCode     *string                `json:"country_code,omitempty"`
	Timezone        *string                `json:"timezone,omitempty"`
	Currency        *string                `json:"currency,omitempty"`
	Website         *string                `json:"website,omitempty"`
	LogoURL         *string                `json:"logo_url,omitempty"`
	Description     *string                `json:"description,omitempty"`
	Status          *string                `json:"status,omitempty"`
	OwnerID         *string                `json:"owner_id,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// [VN] ListCompaniesQuery định nghĩa các tham số lọc, tìm kiếm và phân trang cursor cho danh sách công ty
type ListCompaniesQuery struct {
	Cursor      string `form:"cursor"`
	Limit       int    `form:"limit"`
	Search      string `form:"search"`
	Status      string `form:"status"`
	OwnerID     string `form:"owner_id"`
	CountryCode string `form:"country_code"`
}

// [VN] CompanyOwnerResponse đóng gói thông tin tóm tắt của chủ sở hữu công ty (trường ID mang giá trị public_id để bảo mật)
type CompanyOwnerResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FullName  string  `json:"full_name"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// [VN] CompanyResponse định nghĩa dữ liệu chi tiết của công ty trả về cho client (trường ID mang giá trị public_id để bảo mật)
type CompanyResponse struct {
	ID              string                 `json:"id"`
	Code            string                 `json:"code"`
	Name            string                 `json:"name"`
	LegalName       *string                `json:"legal_name,omitempty"`
	TaxID           *string                `json:"tax_id,omitempty"`
	Email           *string                `json:"email,omitempty"`
	Phone           *string                `json:"phone,omitempty"`
	Address         *string                `json:"address,omitempty"`
	City            *string                `json:"city,omitempty"`
	StateOrProvince *string                `json:"state_or_province,omitempty"`
	PostalCode      *string                `json:"postal_code,omitempty"`
	CountryCode     *string                `json:"country_code,omitempty"`
	Timezone        *string                `json:"timezone,omitempty"`
	Currency        *string                `json:"currency,omitempty"`
	Website         *string                `json:"website,omitempty"`
	LogoURL         *string                `json:"logo_url,omitempty"`
	Description     *string                `json:"description,omitempty"`
	Status          string                 `json:"status"`
	Owner           *CompanyOwnerResponse  `json:"owner,omitempty"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// [VN] ToCompanyResponse chuyển đổi entity models.Company sang DTO CompanyResponse
func ToCompanyResponse(company *models.Company) *CompanyResponse {
	if company == nil {
		return nil
	}

	metadata := map[string]interface{}(company.Metadata)
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	res := &CompanyResponse{
		ID:              company.PublicID,
		Code:            company.Code,
		Name:            company.Name,
		LegalName:       company.LegalName,
		TaxID:           company.TaxID,
		Email:           company.Email,
		Phone:           company.Phone,
		Address:         company.Address,
		City:            company.City,
		StateOrProvince: company.StateOrProvince,
		PostalCode:      company.PostalCode,
		CountryCode:     company.CountryCode,
		Timezone:        company.Timezone,
		Currency:        company.Currency,
		Website:         company.Website,
		LogoURL:         company.LogoURL,
		Description:     company.Description,
		Status:          company.Status,
		Metadata:        metadata,
		CreatedAt:       company.CreatedAt,
		UpdatedAt:       company.UpdatedAt,
	}

	if company.Owner != nil {
		res.Owner = &CompanyOwnerResponse{
			ID:        company.Owner.PublicID,
			Email:     company.Owner.Email,
			FullName:  company.Owner.FullName,
			AvatarURL: company.Owner.AvatarURL,
		}
	}

	return res
}
