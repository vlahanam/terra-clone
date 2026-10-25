package tests

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vlahanam/terra-clone/common"
	"github.com/vlahanam/terra-clone/dto"
	"github.com/vlahanam/terra-clone/models"
	"github.com/vlahanam/terra-clone/services"
	"gorm.io/gorm"
)

// [VN] TestCompanyService_CreateCompany kiểm thử toàn diện nghiệp vụ tạo mới công ty
func TestCompanyService_CreateCompany(t *testing.T) {
	ctx := context.Background()

	t.Run("Success_WithAllFields", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewCompanyService(mockCompanyRepo, mockUserRepo)

		ownerID := uuid.New()
		ownerUser := &models.User{
			ID:       ownerID,
			PublicID: "usr_owner123",
			Email:    "owner@example.com",
			FullName: "Company Owner",
		}

		taxID := "0102030405"
		email := "contact@terra.vn"
		phone := "0243123456"
		address := "123 Terra Boulevard, Hanoi"
		city := "Hanoi"
		state := "Hanoi"
		postalCode := "100000"
		countryCode := "vn" // Sẽ được tự động chuẩn hóa sang in hoa VN
		timezone := "Asia/Ho_Chi_Minh"
		currency := "vnd" // Sẽ được tự động chuẩn hóa sang in hoa VND
		legalName := "Terra Corporation Vietnam Joint Stock Company"
		website := "https://terra.vn"
		logoURL := "https://terra.vn/logo.png"
		desc := "Leading digital services corporation"

		req := &dto.CreateCompanyRequest{
			Name:            "Terra Corporation Vietnam",
			LegalName:       &legalName,
			Code:            "terra_vn", // Sẽ được tự động chuẩn hóa sang in hoa TERRA_VN
			TaxID:           &taxID,
			Email:           &email,
			Phone:           &phone,
			Address:         &address,
			City:            &city,
			StateOrProvince: &state,
			PostalCode:      &postalCode,
			CountryCode:     &countryCode,
			Timezone:        &timezone,
			Currency:        &currency,
			Website:         &website,
			LogoURL:         &logoURL,
			Description:     &desc,
			Status:          models.CompanyStatusActive,
			Metadata: map[string]interface{}{
				"industry": "Technology",
				"size":     "100-500",
			},
		}

		// [VN] Giả lập kiểm tra trùng mã code và mã số thuế
		mockCompanyRepo.On("FindByCode", ctx, "TERRA_VN").Return(nil, gorm.ErrRecordNotFound)
		mockCompanyRepo.On("FindByTaxID", ctx, taxID).Return(nil, gorm.ErrRecordNotFound)
		mockUserRepo.On("FindByID", ctx, ownerID).Return(ownerUser, nil)

		// [VN] Giả lập thao tác tạo mới bản ghi
		mockCompanyRepo.On("Create", ctx, mock.Anything, mock.MatchedBy(func(c *models.Company) bool {
			return c.Code == "TERRA_VN" && c.Name == req.Name && c.OwnerID != nil && *c.OwnerID == ownerID &&
				c.CountryCode != nil && *c.CountryCode == "VN" && c.Currency != nil && *c.Currency == "VND"
		})).Return(nil)

		resp, err := service.CreateCompany(ctx, req, &ownerID)
		require.NoError(t, err, "company creation should succeed")
		require.NotNil(t, resp, "response should not be nil")
		assert.Equal(t, "TERRA_VN", resp.Code, "code should be uppercased")
		assert.Equal(t, req.Name, resp.Name, "name should match")
		assert.Equal(t, &taxID, resp.TaxID, "tax ID should match")
		assert.Equal(t, "VN", *resp.CountryCode, "country code should be uppercased")
		assert.Equal(t, "VND", *resp.Currency, "currency should be uppercased")
		assert.Equal(t, &legalName, resp.LegalName, "legal name should match")
		assert.Equal(t, &city, resp.City, "city should match")
		assert.Equal(t, models.CompanyStatusActive, resp.Status, "status should be active")
		require.NotNil(t, resp.Owner, "owner should be populated")
		assert.Equal(t, ownerUser.PublicID, resp.Owner.ID, "owner public_id should match")
		assert.Equal(t, "Technology", resp.Metadata["industry"], "metadata should match")

		mockCompanyRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Success_WithoutOwner", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewCompanyService(mockCompanyRepo, mockUserRepo)

		req := &dto.CreateCompanyRequest{
			Name: "Independent Startup",
			Code: "STARTUP_XYZ",
		}

		mockCompanyRepo.On("FindByCode", ctx, "STARTUP_XYZ").Return(nil, gorm.ErrRecordNotFound)
		mockCompanyRepo.On("Create", ctx, mock.Anything, mock.MatchedBy(func(c *models.Company) bool {
			return c.Code == "STARTUP_XYZ" && c.OwnerID == nil
		})).Return(nil)

		resp, err := service.CreateCompany(ctx, req, nil)
		require.NoError(t, err, "company creation without owner should succeed")
		require.NotNil(t, resp, "response should not be nil")
		assert.Equal(t, "STARTUP_XYZ", resp.Code)
		assert.Nil(t, resp.Owner, "owner should be nil")

		mockCompanyRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("EmptyCode", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewCompanyService(mockCompanyRepo, mockUserRepo)

		req := &dto.CreateCompanyRequest{
			Name: "Invalid Company",
			Code: "   ",
		}

		resp, err := service.CreateCompany(ctx, req, nil)
		require.Error(t, err, "should fail when code is empty")
		assert.Nil(t, resp)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 400, de.StatusCode())
	})

	t.Run("CodeAlreadyExists", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewCompanyService(mockCompanyRepo, mockUserRepo)

		req := &dto.CreateCompanyRequest{
			Name: "Existing Code Corp",
			Code: "DUPLICATE",
		}

		existing := &models.Company{
			ID:   uuid.New(),
			Code: "DUPLICATE",
			Name: "First Corp",
		}
		mockCompanyRepo.On("FindByCode", ctx, "DUPLICATE").Return(existing, nil)

		resp, err := service.CreateCompany(ctx, req, nil)
		require.Error(t, err, "should fail on duplicate code")
		assert.Nil(t, resp)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 409, de.StatusCode())

		mockCompanyRepo.AssertExpectations(t)
	})

	t.Run("CodeCheckDBError", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewCompanyService(mockCompanyRepo, mockUserRepo)

		req := &dto.CreateCompanyRequest{
			Name: "DB Error Corp",
			Code: "DB_ERR",
		}

		mockCompanyRepo.On("FindByCode", ctx, "DB_ERR").Return(nil, errors.New("db connection failure"))

		resp, err := service.CreateCompany(ctx, req, nil)
		require.Error(t, err, "should fail when DB query fails")
		assert.Nil(t, resp)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 500, de.StatusCode())

		mockCompanyRepo.AssertExpectations(t)
	})

	t.Run("TaxIDAlreadyExists", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewCompanyService(mockCompanyRepo, mockUserRepo)

		taxID := "9999999999"
		req := &dto.CreateCompanyRequest{
			Name:  "Duplicate Tax Corp",
			Code:  "DUP_TAX",
			TaxID: &taxID,
		}

		existingCompany := &models.Company{
			ID:    uuid.New(),
			Code:  "OTHER_CORP",
			TaxID: &taxID,
		}

		mockCompanyRepo.On("FindByCode", ctx, "DUP_TAX").Return(nil, gorm.ErrRecordNotFound)
		mockCompanyRepo.On("FindByTaxID", ctx, taxID).Return(existingCompany, nil)

		resp, err := service.CreateCompany(ctx, req, nil)
		require.Error(t, err, "should fail on duplicate tax id")
		assert.Nil(t, resp)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 409, de.StatusCode())

		mockCompanyRepo.AssertExpectations(t)
	})

	t.Run("OwnerNotFound", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewCompanyService(mockCompanyRepo, mockUserRepo)

		ownerID := uuid.New()
		req := &dto.CreateCompanyRequest{
			Name: "Owner Missing Corp",
			Code: "OWNER_MISS",
		}

		mockCompanyRepo.On("FindByCode", ctx, "OWNER_MISS").Return(nil, gorm.ErrRecordNotFound)
		mockUserRepo.On("FindByID", ctx, ownerID).Return(nil, gorm.ErrRecordNotFound)

		resp, err := service.CreateCompany(ctx, req, &ownerID)
		require.Error(t, err, "should fail when owner is not found")
		assert.Nil(t, resp)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 400, de.StatusCode())

		mockCompanyRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("CreateRecordDBError", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewCompanyService(mockCompanyRepo, mockUserRepo)

		req := &dto.CreateCompanyRequest{
			Name: "Insert Failure Corp",
			Code: "INSERT_FAIL",
		}

		mockCompanyRepo.On("FindByCode", ctx, "INSERT_FAIL").Return(nil, gorm.ErrRecordNotFound)
		mockCompanyRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(errors.New("db disk full"))

		resp, err := service.CreateCompany(ctx, req, nil)
		require.Error(t, err, "should fail on insert error")
		assert.Nil(t, resp)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 500, de.StatusCode())

		mockCompanyRepo.AssertExpectations(t)
	})
}

// [VN] TestCompanyService_GetCompanyByID kiểm thử tra cứu chi tiết công ty theo ID/PublicID/Code
func TestCompanyService_GetCompanyByID(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewCompanyService(mockCompanyRepo, mockUserRepo)

		compID := uuid.New()
		company := &models.Company{
			ID:       compID,
			PublicID: "cmp_123456",
			Code:     "TERRA_HQ",
			Name:     "Terra Headquarters",
			Status:   models.CompanyStatusActive,
		}

		mockCompanyRepo.On("FindByIDOrPublicIDOrCode", ctx, "cmp_123456").Return(company, nil)

		resp, err := service.GetCompanyByID(ctx, "cmp_123456")
		require.NoError(t, err, "should retrieve company successfully")
		require.NotNil(t, resp)
		assert.Equal(t, "cmp_123456", resp.ID)
		assert.Equal(t, "TERRA_HQ", resp.Code)

		// [VN] Kiểm tra JSON serialization đảm bảo trường json:id chứa public_id và không để lộ UUID
		jsonBytes, err := json.Marshal(resp)
		require.NoError(t, err)
		var jsonMap map[string]interface{}
		require.NoError(t, json.Unmarshal(jsonBytes, &jsonMap))
		assert.Equal(t, "cmp_123456", jsonMap["id"], "company response json 'id' field must contain public_id")
		assert.True(t, strings.HasPrefix(jsonMap["id"].(string), "cmp_"), "id field must have cmp_ prefix")

		mockCompanyRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewCompanyService(mockCompanyRepo, mockUserRepo)

		mockCompanyRepo.On("FindByIDOrPublicIDOrCode", ctx, "nonexistent").Return(nil, gorm.ErrRecordNotFound)

		resp, err := service.GetCompanyByID(ctx, "nonexistent")
		require.Error(t, err, "should return error for nonexistent company")
		assert.Nil(t, resp)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 404, de.StatusCode())

		mockCompanyRepo.AssertExpectations(t)
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewCompanyService(mockCompanyRepo, mockUserRepo)

		mockCompanyRepo.On("FindByIDOrPublicIDOrCode", ctx, "cmp_error").Return(nil, errors.New("query failed"))

		resp, err := service.GetCompanyByID(ctx, "cmp_error")
		require.Error(t, err, "should return error when DB fails")
		assert.Nil(t, resp)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 500, de.StatusCode())

		mockCompanyRepo.AssertExpectations(t)
	})
}

// [VN] TestCompanyService_ListCompanies kiểm thử danh sách công ty có phân trang cursor và lọc
func TestCompanyService_ListCompanies(t *testing.T) {
	ctx := context.Background()

	t.Run("Success_Default", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewCompanyService(mockCompanyRepo, mockUserRepo)

		comp1 := models.Company{
			ID:        uuid.New(),
			PublicID:  "cmp_aaa",
			Code:      "AAA",
			Name:      "Alpha Corp",
			Status:    models.CompanyStatusActive,
			CreatedAt: time.Now(),
		}
		comp2 := models.Company{
			ID:        uuid.New(),
			PublicID:  "cmp_bbb",
			Code:      "BBB",
			Name:      "Beta Corp",
			Status:    models.CompanyStatusActive,
			CreatedAt: time.Now(),
		}

		query := &dto.ListCompaniesQuery{
			Limit: 10,
		}

		mockCompanyRepo.On("List", ctx, "", 10, "", "", (*uuid.UUID)(nil), "").
			Return([]models.Company{comp1, comp2}, comp2.PublicID, int64(2), nil)

		res, paging, err := service.ListCompanies(ctx, query)
		require.NoError(t, err, "listing companies should succeed")
		require.Len(t, res, 2)
		assert.Equal(t, "cmp_aaa", res[0].ID)
		assert.Equal(t, "cmp_bbb", res[1].ID)
		require.NotNil(t, paging)
		assert.Equal(t, int64(2), paging.Total)
		assert.Equal(t, comp2.PublicID, paging.NextCursor)

		mockCompanyRepo.AssertExpectations(t)
	})

	t.Run("Success_WithCountryFilter", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewCompanyService(mockCompanyRepo, mockUserRepo)

		vnComp := models.Company{
			ID:       uuid.New(),
			PublicID: "cmp_vn1",
			Code:     "VN_CORP",
			Name:     "Vietnam Corp",
		}

		query := &dto.ListCompaniesQuery{
			CountryCode: "VN",
			Limit:       10,
		}

		mockCompanyRepo.On("List", ctx, "", 10, "", "", (*uuid.UUID)(nil), "VN").
			Return([]models.Company{vnComp}, "", int64(1), nil)

		res, paging, err := service.ListCompanies(ctx, query)
		require.NoError(t, err)
		require.Len(t, res, 1)
		assert.Equal(t, "cmp_vn1", res[0].ID)
		assert.Equal(t, int64(1), paging.Total)

		mockCompanyRepo.AssertExpectations(t)
	})

	t.Run("Success_WithOwnerFilter", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewCompanyService(mockCompanyRepo, mockUserRepo)

		ownerID := uuid.New()
		ownerUser := &models.User{
			ID:       ownerID,
			PublicID: "usr_owner_filter",
		}

		query := &dto.ListCompaniesQuery{
			OwnerID: "usr_owner_filter",
			Limit:   5,
		}

		mockUserRepo.On("FindByIDOrPublicID", ctx, "usr_owner_filter").Return(ownerUser, nil)
		mockCompanyRepo.On("List", ctx, "", 5, "", "", &ownerID, "").
			Return([]models.Company{}, "", int64(0), nil)

		res, paging, err := service.ListCompanies(ctx, query)
		require.NoError(t, err)
		assert.Empty(t, res)
		assert.Equal(t, int64(0), paging.Total)

		mockCompanyRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewCompanyService(mockCompanyRepo, mockUserRepo)

		query := &dto.ListCompaniesQuery{Limit: 10}
		mockCompanyRepo.On("List", ctx, "", 10, "", "", (*uuid.UUID)(nil), "").
			Return(nil, "", int64(0), errors.New("list query failed"))

		res, paging, err := service.ListCompanies(ctx, query)
		require.Error(t, err)
		assert.Nil(t, res)
		assert.Nil(t, paging)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 500, de.StatusCode())

		mockCompanyRepo.AssertExpectations(t)
	})
}

// [VN] TestCompanyService_UpdateCompany kiểm thử cập nhật thông tin công ty và xử lý xung đột
func TestCompanyService_UpdateCompany(t *testing.T) {
	ctx := context.Background()

	t.Run("Success_UpdateBasic", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewCompanyService(mockCompanyRepo, mockUserRepo)

		compID := uuid.New()
		company := &models.Company{
			ID:       compID,
			PublicID: "cmp_update1",
			Code:     "ORIGINAL_CODE",
			Name:     "Original Name",
			Status:   models.CompanyStatusActive,
			Metadata: models.JSONMap{"old_key": "old_val"},
		}

		newName := "Updated Corp Name"
		newWebsite := "https://updated.corp"
		newStatus := models.CompanyStatusInactive
		newCountry := "sg" // Sẽ được tự động chuẩn hóa in hoa SG
		newCurrency := "sgd" // Sẽ được tự động chuẩn hóa in hoa SGD
		newCity := "Singapore"
		newLegalName := "Updated Corp Pte. Ltd."
		req := &dto.UpdateCompanyRequest{
			Name:            &newName,
			LegalName:       &newLegalName,
			City:            &newCity,
			CountryCode:     &newCountry,
			Currency:        &newCurrency,
			Website:         &newWebsite,
			Status:          &newStatus,
			Metadata: map[string]interface{}{
				"new_key": "new_val",
			},
		}

		mockCompanyRepo.On("FindByIDOrPublicIDOrCode", ctx, "cmp_update1").Return(company, nil)
		mockCompanyRepo.On("Update", ctx, mock.Anything, mock.MatchedBy(func(c *models.Company) bool {
			return c.Name == newName && *c.Website == newWebsite && c.Status == newStatus && c.Metadata["new_key"] == "new_val" &&
				c.CountryCode != nil && *c.CountryCode == "SG" && c.Currency != nil && *c.Currency == "SGD" &&
				c.LegalName != nil && *c.LegalName == newLegalName && c.City != nil && *c.City == newCity
		})).Return(nil)

		resp, err := service.UpdateCompany(ctx, "cmp_update1", req)
		require.NoError(t, err, "update company should succeed")
		require.NotNil(t, resp)
		assert.Equal(t, newName, resp.Name)
		assert.Equal(t, &newLegalName, resp.LegalName)
		assert.Equal(t, &newCity, resp.City)
		assert.Equal(t, "SG", *resp.CountryCode)
		assert.Equal(t, "SGD", *resp.Currency)
		assert.Equal(t, &newWebsite, resp.Website)
		assert.Equal(t, newStatus, resp.Status)
		assert.Equal(t, "new_val", resp.Metadata["new_key"])

		mockCompanyRepo.AssertExpectations(t)
	})

	t.Run("Success_UpdateCode_NoConflict", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewCompanyService(mockCompanyRepo, mockUserRepo)

		compID := uuid.New()
		company := &models.Company{
			ID:       compID,
			PublicID: "cmp_update_code",
			Code:     "OLD_CODE",
			Name:     "Code Change Corp",
		}

		newCode := "new_code" // Sẽ được tự động chuẩn hóa sang in hoa NEW_CODE
		req := &dto.UpdateCompanyRequest{
			Code: &newCode,
		}

		mockCompanyRepo.On("FindByIDOrPublicIDOrCode", ctx, "cmp_update_code").Return(company, nil)
		mockCompanyRepo.On("FindByCode", ctx, "NEW_CODE").Return(nil, gorm.ErrRecordNotFound)
		mockCompanyRepo.On("Update", ctx, mock.Anything, mock.MatchedBy(func(c *models.Company) bool {
			return c.Code == "NEW_CODE"
		})).Return(nil)

		resp, err := service.UpdateCompany(ctx, "cmp_update_code", req)
		require.NoError(t, err)
		assert.Equal(t, "NEW_CODE", resp.Code)

		mockCompanyRepo.AssertExpectations(t)
	})

	t.Run("CodeConflict", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewCompanyService(mockCompanyRepo, mockUserRepo)

		compID := uuid.New()
		company := &models.Company{
			ID:       compID,
			PublicID: "cmp_conflict",
			Code:     "MY_CODE",
		}

		otherCompID := uuid.New()
		otherCompany := &models.Company{
			ID:   otherCompID,
			Code: "TAKEN_CODE",
		}

		newCode := "TAKEN_CODE"
		req := &dto.UpdateCompanyRequest{
			Code: &newCode,
		}

		mockCompanyRepo.On("FindByIDOrPublicIDOrCode", ctx, "cmp_conflict").Return(company, nil)
		mockCompanyRepo.On("FindByCode", ctx, "TAKEN_CODE").Return(otherCompany, nil)

		resp, err := service.UpdateCompany(ctx, "cmp_conflict", req)
		require.Error(t, err, "should fail on code conflict with other company")
		assert.Nil(t, resp)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 409, de.StatusCode())

		mockCompanyRepo.AssertExpectations(t)
	})

	t.Run("TaxIDConflict", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewCompanyService(mockCompanyRepo, mockUserRepo)

		compID := uuid.New()
		company := &models.Company{
			ID:       compID,
			PublicID: "cmp_tax_conflict",
			Code:     "MY_CORP",
		}

		otherCompID := uuid.New()
		otherCompany := &models.Company{
			ID:   otherCompID,
			Code: "OTHER_CORP",
		}

		newTaxID := "8888888888"
		req := &dto.UpdateCompanyRequest{
			TaxID: &newTaxID,
		}

		mockCompanyRepo.On("FindByIDOrPublicIDOrCode", ctx, "cmp_tax_conflict").Return(company, nil)
		mockCompanyRepo.On("FindByTaxID", ctx, newTaxID).Return(otherCompany, nil)

		resp, err := service.UpdateCompany(ctx, "cmp_tax_conflict", req)
		require.Error(t, err, "should fail on tax id conflict")
		assert.Nil(t, resp)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 409, de.StatusCode())

		mockCompanyRepo.AssertExpectations(t)
	})

	t.Run("UpdateOwner_Success", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewCompanyService(mockCompanyRepo, mockUserRepo)

		compID := uuid.New()
		company := &models.Company{
			ID:       compID,
			PublicID: "cmp_owner_update",
			Code:     "OWNER_TEST",
		}

		newOwnerUUID := uuid.New()
		newOwner := &models.User{
			ID:       newOwnerUUID,
			PublicID: "usr_new_owner",
			FullName: "New Owner",
		}

		newOwnerStr := "usr_new_owner"
		req := &dto.UpdateCompanyRequest{
			OwnerID: &newOwnerStr,
		}

		mockCompanyRepo.On("FindByIDOrPublicIDOrCode", ctx, "cmp_owner_update").Return(company, nil)
		mockUserRepo.On("FindByIDOrPublicID", ctx, "usr_new_owner").Return(newOwner, nil)
		mockCompanyRepo.On("Update", ctx, mock.Anything, mock.MatchedBy(func(c *models.Company) bool {
			return c.OwnerID != nil && *c.OwnerID == newOwnerUUID
		})).Return(nil)

		resp, err := service.UpdateCompany(ctx, "cmp_owner_update", req)
		require.NoError(t, err)
		require.NotNil(t, resp.Owner)
		assert.Equal(t, "usr_new_owner", resp.Owner.ID)

		mockCompanyRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("UpdateOwner_Remove", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewCompanyService(mockCompanyRepo, mockUserRepo)

		compID := uuid.New()
		ownerUUID := uuid.New()
		company := &models.Company{
			ID:       compID,
			PublicID: "cmp_remove_owner",
			Code:     "REMOVE_OWNER",
			OwnerID:  &ownerUUID,
		}

		emptyOwner := ""
		req := &dto.UpdateCompanyRequest{
			OwnerID: &emptyOwner,
		}

		mockCompanyRepo.On("FindByIDOrPublicIDOrCode", ctx, "cmp_remove_owner").Return(company, nil)
		mockCompanyRepo.On("Update", ctx, mock.Anything, mock.MatchedBy(func(c *models.Company) bool {
			return c.OwnerID == nil && c.Owner == nil
		})).Return(nil)

		resp, err := service.UpdateCompany(ctx, "cmp_remove_owner", req)
		require.NoError(t, err)
		assert.Nil(t, resp.Owner)

		mockCompanyRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewCompanyService(mockCompanyRepo, mockUserRepo)

		name := "Name"
		req := &dto.UpdateCompanyRequest{Name: &name}

		mockCompanyRepo.On("FindByIDOrPublicIDOrCode", ctx, "not_exist").Return(nil, gorm.ErrRecordNotFound)

		resp, err := service.UpdateCompany(ctx, "not_exist", req)
		require.Error(t, err)
		assert.Nil(t, resp)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 404, de.StatusCode())

		mockCompanyRepo.AssertExpectations(t)
	})
}

// [VN] TestCompanyService_DeleteCompany kiểm thử xóa mềm và xóa cứng công ty
func TestCompanyService_DeleteCompany(t *testing.T) {
	ctx := context.Background()

	t.Run("SoftDelete_Success", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewCompanyService(mockCompanyRepo, mockUserRepo)

		compID := uuid.New()
		company := &models.Company{
			ID:       compID,
			PublicID: "cmp_soft_del",
			Status:   models.CompanyStatusActive,
		}

		mockCompanyRepo.On("FindByIDOrPublicIDOrCode", ctx, "cmp_soft_del").Return(company, nil)
		mockCompanyRepo.On("Delete", ctx, mock.Anything, compID, true).Return(nil)

		err := service.DeleteCompany(ctx, "cmp_soft_del", false)
		require.NoError(t, err, "soft delete should succeed")

		mockCompanyRepo.AssertExpectations(t)
	})

	t.Run("HardDelete_Success", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewCompanyService(mockCompanyRepo, mockUserRepo)

		compID := uuid.New()
		company := &models.Company{
			ID:       compID,
			PublicID: "cmp_hard_del",
			Status:   models.CompanyStatusActive,
		}

		mockCompanyRepo.On("FindByIDOrPublicIDOrCode", ctx, "cmp_hard_del").Return(company, nil)
		mockCompanyRepo.On("Delete", ctx, mock.Anything, compID, false).Return(nil)

		err := service.DeleteCompany(ctx, "cmp_hard_del", true)
		require.NoError(t, err, "hard delete should succeed")

		mockCompanyRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewCompanyService(mockCompanyRepo, mockUserRepo)

		mockCompanyRepo.On("FindByIDOrPublicIDOrCode", ctx, "missing").Return(nil, gorm.ErrRecordNotFound)

		err := service.DeleteCompany(ctx, "missing", false)
		require.Error(t, err, "should return error for missing company")
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 404, de.StatusCode())

		mockCompanyRepo.AssertExpectations(t)
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewCompanyService(mockCompanyRepo, mockUserRepo)

		compID := uuid.New()
		company := &models.Company{
			ID:       compID,
			PublicID: "cmp_err_del",
		}

		mockCompanyRepo.On("FindByIDOrPublicIDOrCode", ctx, "cmp_err_del").Return(company, nil)
		mockCompanyRepo.On("Delete", ctx, mock.Anything, compID, true).Return(errors.New("db delete error"))

		err := service.DeleteCompany(ctx, "cmp_err_del", false)
		require.Error(t, err, "should return error on DB delete failure")
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 500, de.StatusCode())

		mockCompanyRepo.AssertExpectations(t)
	})
}
