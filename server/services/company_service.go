package services

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/vlahanam/terra-clone/common"
	"github.com/vlahanam/terra-clone/dto"
	"github.com/vlahanam/terra-clone/models"
	"github.com/vlahanam/terra-clone/repositories"
	"gorm.io/gorm"
)

// [VN] CompanyService định nghĩa các phương thức xử lý nghiệp vụ quản lý công ty
type CompanyService interface {
	CreateCompany(ctx context.Context, req *dto.CreateCompanyRequest, ownerID *uuid.UUID) (*dto.CompanyResponse, error)
	GetCompanyByID(ctx context.Context, identifier string) (*dto.CompanyResponse, error)
	ListCompanies(ctx context.Context, query *dto.ListCompaniesQuery) ([]dto.CompanyResponse, *common.Paging, error)
	UpdateCompany(ctx context.Context, identifier string, req *dto.UpdateCompanyRequest) (*dto.CompanyResponse, error)
	DeleteCompany(ctx context.Context, identifier string, hard bool) error
}

type companyService struct {
	companyRepo repositories.CompanyRepository
	userRepo    repositories.UserRepository
}

// [VN] NewCompanyService khởi tạo CompanyService với CompanyRepository và UserRepository
func NewCompanyService(
	companyRepo repositories.CompanyRepository,
	userRepo repositories.UserRepository,
) CompanyService {
	return &companyService{
		companyRepo: companyRepo,
		userRepo:    userRepo,
	}
}

// [VN] CreateCompany tạo mới công ty sau khi kiểm tra tính hợp lệ và duy nhất của mã code và mã số thuế
func (s *companyService) CreateCompany(
	ctx context.Context,
	req *dto.CreateCompanyRequest,
	ownerID *uuid.UUID,
) (*dto.CompanyResponse, error) {
	cleanCode := strings.ToUpper(strings.TrimSpace(req.Code))
	if cleanCode == "" {
		return nil, common.ErrBadRequest.
			WithReason("Company code cannot be empty").
			WithDetail("code", req.Code)
	}

	// [VN] Kiểm tra trùng lặp mã code của công ty
	existingCode, err := s.companyRepo.FindByCode(ctx, cleanCode)
	if err == nil && existingCode != nil {
		return nil, common.ErrConflict.
			WithReason("Company code is already in use in the system").
			WithDetail("code", cleanCode)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.ErrInternalServerError.
			WithReason("Failed to verify company code uniqueness").
			WithTrace(err)
	}

	// [VN] Kiểm tra trùng mã số thuế nếu có cung cấp
	if req.TaxID != nil && strings.TrimSpace(*req.TaxID) != "" {
		cleanTaxID := strings.TrimSpace(*req.TaxID)
		existingTax, err := s.companyRepo.FindByTaxID(ctx, cleanTaxID)
		if err == nil && existingTax != nil {
			return nil, common.ErrConflict.
				WithReason("Tax ID is already registered to another company").
				WithDetail("tax_id", cleanTaxID)
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrInternalServerError.
				WithReason("Failed to verify company tax ID uniqueness").
				WithTrace(err)
		}
	}

	var ownerUser *models.User
	if ownerID != nil {
		owner, err := s.userRepo.FindByID(ctx, *ownerID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, common.ErrBadRequest.
					WithReason("Owner user not found in the system").
					WithDetail("owner", "current_user")
			}
			return nil, common.ErrInternalServerError.
				WithReason("Failed to verify company owner").
				WithTrace(err)
		}
		ownerUser = owner
	}

	status := models.CompanyStatusActive
	if req.Status != "" {
		status = req.Status
	}

	metadata := models.JSONMap(req.Metadata)
	if metadata == nil {
		metadata = make(models.JSONMap)
	}

	var countryCode *string
	if req.CountryCode != nil && strings.TrimSpace(*req.CountryCode) != "" {
		code := strings.ToUpper(strings.TrimSpace(*req.CountryCode))
		countryCode = &code
	}

	var currency *string
	if req.Currency != nil && strings.TrimSpace(*req.Currency) != "" {
		curr := strings.ToUpper(strings.TrimSpace(*req.Currency))
		currency = &curr
	}

	newCompany := &models.Company{
		PublicID:        common.GeneratePublicID(common.PREFIX_COMPANY),
		Code:            cleanCode,
		Name:            strings.TrimSpace(req.Name),
		LegalName:       req.LegalName,
		TaxID:           req.TaxID,
		Email:           req.Email,
		Phone:           req.Phone,
		Address:         req.Address,
		City:            req.City,
		StateOrProvince: req.StateOrProvince,
		PostalCode:      req.PostalCode,
		CountryCode:     countryCode,
		Timezone:        req.Timezone,
		Currency:        currency,
		Website:         req.Website,
		LogoURL:         req.LogoURL,
		Description:     req.Description,
		Status:          status,
		OwnerID:         ownerID,
		Owner:           ownerUser,
		Metadata:        metadata,
	}

	if err := s.companyRepo.Create(ctx, nil, newCompany); err != nil {
		return nil, common.ErrInternalServerError.
			WithReason("Failed to create company record").
			WithTrace(err)
	}

	return dto.ToCompanyResponse(newCompany), nil
}

// [VN] GetCompanyByID lấy thông tin chi tiết công ty theo ID, PublicID hoặc Code
func (s *companyService) GetCompanyByID(ctx context.Context, identifier string) (*dto.CompanyResponse, error) {
	company, err := s.companyRepo.FindByIDOrPublicIDOrCode(ctx, identifier)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound.
				WithReason("Company does not exist in the system").
				WithDetail("identifier", identifier)
		}
		return nil, common.ErrInternalServerError.
			WithReason("Failed to query company details").
			WithTrace(err)
	}

	return dto.ToCompanyResponse(company), nil
}

// [VN] ListCompanies trả về danh sách công ty kèm phân trang cursor và bộ lọc
func (s *companyService) ListCompanies(
	ctx context.Context,
	query *dto.ListCompaniesQuery,
) ([]dto.CompanyResponse, *common.Paging, error) {
	var ownerUUID *uuid.UUID
	if query.OwnerID != "" {
		user, err := s.userRepo.FindByIDOrPublicID(ctx, query.OwnerID)
		if err == nil && user != nil {
			ownerUUID = &user.ID
		} else {
			if parsed, err := uuid.Parse(query.OwnerID); err == nil {
				ownerUUID = &parsed
			}
		}
	}

	companies, nextCursor, total, err := s.companyRepo.List(
		ctx,
		query.Cursor,
		query.Limit,
		query.Search,
		query.Status,
		ownerUUID,
		query.CountryCode,
	)
	if err != nil {
		return nil, nil, common.ErrInternalServerError.
			WithReason("Failed to list companies").
			WithTrace(err)
	}

	res := make([]dto.CompanyResponse, 0, len(companies))
	for i := range companies {
		if resp := dto.ToCompanyResponse(&companies[i]); resp != nil {
			res = append(res, *resp)
		}
	}

	paging := &common.Paging{
		Total:      total,
		FakeCursor: query.Cursor,
		NextCursor: nextCursor,
	}

	return res, paging, nil
}

// [VN] UpdateCompany cập nhật thông tin công ty và cấu hình metadata
func (s *companyService) UpdateCompany(
	ctx context.Context,
	identifier string,
	req *dto.UpdateCompanyRequest,
) (*dto.CompanyResponse, error) {
	company, err := s.companyRepo.FindByIDOrPublicIDOrCode(ctx, identifier)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound.
				WithReason("Company does not exist in the system").
				WithDetail("identifier", identifier)
		}
		return nil, common.ErrInternalServerError.
			WithReason("Failed to retrieve company for update").
			WithTrace(err)
	}

	// [VN] Nếu cập nhật mã code, kiểm tra không bị trùng với công ty khác
	if req.Code != nil {
		cleanCode := strings.ToUpper(strings.TrimSpace(*req.Code))
		if cleanCode != company.Code {
			existing, err := s.companyRepo.FindByCode(ctx, cleanCode)
			if err == nil && existing != nil && existing.ID != company.ID {
				return nil, common.ErrConflict.
					WithReason("New company code is already in use").
					WithDetail("code", cleanCode)
			} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, common.ErrInternalServerError.
					WithReason("Failed to verify company code uniqueness").
					WithTrace(err)
			}
			company.Code = cleanCode
		}
	}

	// [VN] Nếu cập nhật mã số thuế, kiểm tra không bị trùng với công ty khác
	if req.TaxID != nil && strings.TrimSpace(*req.TaxID) != "" {
		cleanTax := strings.TrimSpace(*req.TaxID)
		if company.TaxID == nil || cleanTax != *company.TaxID {
			existing, err := s.companyRepo.FindByTaxID(ctx, cleanTax)
			if err == nil && existing != nil && existing.ID != company.ID {
				return nil, common.ErrConflict.
					WithReason("New tax ID is already registered to another company").
					WithDetail("tax_id", cleanTax)
			} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, common.ErrInternalServerError.
					WithReason("Failed to verify tax ID uniqueness").
					WithTrace(err)
			}
			company.TaxID = &cleanTax
		}
	}

	if req.Name != nil {
		company.Name = strings.TrimSpace(*req.Name)
	}
	if req.LegalName != nil {
		company.LegalName = req.LegalName
	}
	if req.Email != nil {
		company.Email = req.Email
	}
	if req.Phone != nil {
		company.Phone = req.Phone
	}
	if req.Address != nil {
		company.Address = req.Address
	}
	if req.City != nil {
		company.City = req.City
	}
	if req.StateOrProvince != nil {
		company.StateOrProvince = req.StateOrProvince
	}
	if req.PostalCode != nil {
		company.PostalCode = req.PostalCode
	}
	if req.CountryCode != nil {
		if strings.TrimSpace(*req.CountryCode) == "" {
			company.CountryCode = nil
		} else {
			code := strings.ToUpper(strings.TrimSpace(*req.CountryCode))
			company.CountryCode = &code
		}
	}
	if req.Timezone != nil {
		company.Timezone = req.Timezone
	}
	if req.Currency != nil {
		if strings.TrimSpace(*req.Currency) == "" {
			company.Currency = nil
		} else {
			curr := strings.ToUpper(strings.TrimSpace(*req.Currency))
			company.Currency = &curr
		}
	}
	if req.Website != nil {
		company.Website = req.Website
	}
	if req.LogoURL != nil {
		company.LogoURL = req.LogoURL
	}
	if req.Description != nil {
		company.Description = req.Description
	}
	if req.Status != nil {
		company.Status = *req.Status
	}

	if req.OwnerID != nil {
		if *req.OwnerID == "" {
			company.OwnerID = nil
			company.Owner = nil
		} else {
			user, err := s.userRepo.FindByIDOrPublicID(ctx, *req.OwnerID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, common.ErrBadRequest.
						WithReason("Owner user not found in the system").
						WithDetail("owner_id", *req.OwnerID)
				}
				return nil, common.ErrInternalServerError.
					WithReason("Failed to verify company owner").
					WithTrace(err)
			}
			company.OwnerID = &user.ID
			company.Owner = user
		}
	}

	// [VN] Cập nhật hoặc hợp nhất các trường cấu hình metadata
	if req.Metadata != nil {
		if company.Metadata == nil {
			company.Metadata = make(models.JSONMap)
		}
		for k, v := range req.Metadata {
			company.Metadata[k] = v
		}
	}

	if err := s.companyRepo.Update(ctx, nil, company); err != nil {
		return nil, common.ErrInternalServerError.
			WithReason("Failed to update company record").
			WithTrace(err)
	}

	return dto.ToCompanyResponse(company), nil
}

// [VN] DeleteCompany xử lý xóa mềm hoặc xóa cứng công ty
func (s *companyService) DeleteCompany(ctx context.Context, identifier string, hard bool) error {
	company, err := s.companyRepo.FindByIDOrPublicIDOrCode(ctx, identifier)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrNotFound.
				WithReason("Company does not exist in the system").
				WithDetail("identifier", identifier)
		}
		return common.ErrInternalServerError.
			WithReason("Failed to retrieve company for deletion").
			WithTrace(err)
	}

	if err := s.companyRepo.Delete(ctx, nil, company.ID, !hard); err != nil {
		return common.ErrInternalServerError.
			WithReason("Failed to delete company record").
			WithTrace(err)
	}

	return nil
}
