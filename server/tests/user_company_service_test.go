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

// [VN] TestUserCompanyService_CreateUserCompany kiểm thử toàn diện nghiệp vụ gán người dùng vào công ty
func TestUserCompanyService_CreateUserCompany(t *testing.T) {
	ctx := context.Background()

	t.Run("Success_NewAssociation", func(t *testing.T) {
		mockUCRepo := new(MockUserCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		mockCompanyRepo := new(MockCompanyRepository)
		service := services.NewUserCompanyService(mockUCRepo, mockUserRepo, mockCompanyRepo)

		userID := uuid.New()
		user := &models.User{
			ID:       userID,
			PublicID: "usr_user123",
			Email:    "user@example.com",
			FullName: "John Doe",
			Status:   models.UserStatusActive,
		}

		compID := uuid.New()
		comp := &models.Company{
			ID:       compID,
			PublicID: "cmp_terra123",
			Code:     "TERRA_VN",
			Name:     "Terra Vietnam",
			Status:   models.CompanyStatusActive,
		}

		jobTitle := "Senior Backend Engineer"
		dept := "Engineering"
		req := &dto.CreateUserCompanyRequest{
			UserID:     user.PublicID,
			CompanyID:  comp.PublicID,
			Role:       "admin",
			JobTitle:   &jobTitle,
			Department: &dept,
			Status:     "active",
			Permissions: map[string]interface{}{
				"can_invite": true,
			},
			Metadata: map[string]interface{}{
				"location": "Hanoi",
			},
		}

		mockUserRepo.On("FindByIDOrPublicID", ctx, user.PublicID).Return(user, nil)
		mockCompanyRepo.On("FindByIDOrPublicIDOrCode", ctx, comp.PublicID).Return(comp, nil)
		mockUCRepo.On("FindByUserAndCompany", ctx, user.ID, comp.ID).Return(nil, gorm.ErrRecordNotFound)
		mockUCRepo.On("Create", ctx, (*gorm.DB)(nil), mock.MatchedBy(func(uc *models.UserCompany) bool {
			return uc.UserID == user.ID && uc.CompanyID == comp.ID && uc.Role == "admin"
		})).Return(nil)

		res, err := service.CreateUserCompany(ctx, req)
		require.NoError(t, err, "should create user company successfully")
		require.NotNil(t, res)
		assert.Equal(t, "admin", res.Role)
		assert.Equal(t, &jobTitle, res.JobTitle)
		assert.Equal(t, &dept, res.Department)
		assert.Equal(t, "active", res.Status)
		assert.NotNil(t, res.User)
		assert.Equal(t, user.PublicID, res.User.ID)
		assert.Equal(t, user.PublicID, res.UserID)
		assert.NotNil(t, res.Company)
		assert.Equal(t, comp.PublicID, res.Company.ID)
		assert.Equal(t, comp.PublicID, res.CompanyID)
		mockUserRepo.AssertExpectations(t)
		mockCompanyRepo.AssertExpectations(t)
		mockUCRepo.AssertExpectations(t)
	})

	t.Run("Success_ReactivateDeleted", func(t *testing.T) {
		mockUCRepo := new(MockUserCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		mockCompanyRepo := new(MockCompanyRepository)
		service := services.NewUserCompanyService(mockUCRepo, mockUserRepo, mockCompanyRepo)

		userID := uuid.New()
		user := &models.User{
			ID:       userID,
			PublicID: "usr_user123",
			Status:   models.UserStatusActive,
		}

		compID := uuid.New()
		comp := &models.Company{
			ID:       compID,
			PublicID: "cmp_terra123",
			Status:   models.CompanyStatusActive,
		}

		existingDeleted := &models.UserCompany{
			ID:        uuid.New(),
			PublicID:  "ucmp_old",
			UserID:    userID,
			CompanyID: compID,
			Role:      "member",
			Status:    models.UserCompanyStatusDeleted,
		}

		req := &dto.CreateUserCompanyRequest{
			UserID:    user.PublicID,
			CompanyID: comp.PublicID,
			Role:      "member",
			Status:    "active",
		}

		mockUserRepo.On("FindByIDOrPublicID", ctx, user.PublicID).Return(user, nil)
		mockCompanyRepo.On("FindByIDOrPublicIDOrCode", ctx, comp.PublicID).Return(comp, nil)
		mockUCRepo.On("FindByUserAndCompany", ctx, user.ID, comp.ID).Return(existingDeleted, nil)
		mockUCRepo.On("Update", ctx, (*gorm.DB)(nil), mock.MatchedBy(func(uc *models.UserCompany) bool {
			return uc.ID == existingDeleted.ID && uc.Status == "active"
		})).Return(nil)

		res, err := service.CreateUserCompany(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "active", res.Status)
		mockUCRepo.AssertExpectations(t)
	})

	t.Run("EmptyUserID", func(t *testing.T) {
		service := services.NewUserCompanyService(nil, nil, nil)
		_, err := service.CreateUserCompany(ctx, &dto.CreateUserCompanyRequest{
			UserID:    "   ",
			CompanyID: "cmp_123",
		})
		require.Error(t, err)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 400, de.StatusCode())
	})

	t.Run("EmptyCompanyID", func(t *testing.T) {
		service := services.NewUserCompanyService(nil, nil, nil)
		_, err := service.CreateUserCompany(ctx, &dto.CreateUserCompanyRequest{
			UserID:    "usr_123",
			CompanyID: "",
		})
		require.Error(t, err)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 400, de.StatusCode())
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		service := services.NewUserCompanyService(nil, mockUserRepo, nil)

		mockUserRepo.On("FindByIDOrPublicID", ctx, "usr_missing").Return(nil, gorm.ErrRecordNotFound)

		_, err := service.CreateUserCompany(ctx, &dto.CreateUserCompanyRequest{
			UserID:    "usr_missing",
			CompanyID: "cmp_123",
		})
		require.Error(t, err)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 404, de.StatusCode())
	})

	t.Run("UserDeleted", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		service := services.NewUserCompanyService(nil, mockUserRepo, nil)

		user := &models.User{
			ID:       uuid.New(),
			PublicID: "usr_deleted",
			Status:   models.UserStatusDeleted,
		}
		mockUserRepo.On("FindByIDOrPublicID", ctx, user.PublicID).Return(user, nil)

		_, err := service.CreateUserCompany(ctx, &dto.CreateUserCompanyRequest{
			UserID:    user.PublicID,
			CompanyID: "cmp_123",
		})
		require.Error(t, err)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 400, de.StatusCode())
	})

	t.Run("UserDBError", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		service := services.NewUserCompanyService(nil, mockUserRepo, nil)

		mockUserRepo.On("FindByIDOrPublicID", ctx, "usr_123").Return(nil, errors.New("db error"))

		_, err := service.CreateUserCompany(ctx, &dto.CreateUserCompanyRequest{
			UserID:    "usr_123",
			CompanyID: "cmp_123",
		})
		require.Error(t, err)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 500, de.StatusCode())
	})

	t.Run("CompanyNotFound", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCompanyRepo := new(MockCompanyRepository)
		service := services.NewUserCompanyService(nil, mockUserRepo, mockCompanyRepo)

		user := &models.User{
			ID:       uuid.New(),
			PublicID: "usr_active",
			Status:   models.UserStatusActive,
		}
		mockUserRepo.On("FindByIDOrPublicID", ctx, user.PublicID).Return(user, nil)
		mockCompanyRepo.On("FindByIDOrPublicIDOrCode", ctx, "cmp_missing").Return(nil, gorm.ErrRecordNotFound)

		_, err := service.CreateUserCompany(ctx, &dto.CreateUserCompanyRequest{
			UserID:    user.PublicID,
			CompanyID: "cmp_missing",
		})
		require.Error(t, err)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 404, de.StatusCode())
	})

	t.Run("CompanyDeleted", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCompanyRepo := new(MockCompanyRepository)
		service := services.NewUserCompanyService(nil, mockUserRepo, mockCompanyRepo)

		user := &models.User{
			ID:       uuid.New(),
			PublicID: "usr_active",
			Status:   models.UserStatusActive,
		}
		comp := &models.Company{
			ID:       uuid.New(),
			PublicID: "cmp_deleted",
			Status:   models.CompanyStatusDeleted,
		}
		mockUserRepo.On("FindByIDOrPublicID", ctx, user.PublicID).Return(user, nil)
		mockCompanyRepo.On("FindByIDOrPublicIDOrCode", ctx, comp.PublicID).Return(comp, nil)

		_, err := service.CreateUserCompany(ctx, &dto.CreateUserCompanyRequest{
			UserID:    user.PublicID,
			CompanyID: comp.PublicID,
		})
		require.Error(t, err)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 400, de.StatusCode())
	})

	t.Run("CompanyDBError", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCompanyRepo := new(MockCompanyRepository)
		service := services.NewUserCompanyService(nil, mockUserRepo, mockCompanyRepo)

		user := &models.User{
			ID:       uuid.New(),
			PublicID: "usr_active",
			Status:   models.UserStatusActive,
		}
		mockUserRepo.On("FindByIDOrPublicID", ctx, user.PublicID).Return(user, nil)
		mockCompanyRepo.On("FindByIDOrPublicIDOrCode", ctx, "cmp_123").Return(nil, errors.New("db error"))

		_, err := service.CreateUserCompany(ctx, &dto.CreateUserCompanyRequest{
			UserID:    user.PublicID,
			CompanyID: "cmp_123",
		})
		require.Error(t, err)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 500, de.StatusCode())
	})

	t.Run("InvalidRole", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCompanyRepo := new(MockCompanyRepository)
		service := services.NewUserCompanyService(nil, mockUserRepo, mockCompanyRepo)

		user := &models.User{
			ID:       uuid.New(),
			PublicID: "usr_active",
			Status:   models.UserStatusActive,
		}
		comp := &models.Company{
			ID:       uuid.New(),
			PublicID: "cmp_active",
			Status:   models.CompanyStatusActive,
		}
		mockUserRepo.On("FindByIDOrPublicID", ctx, user.PublicID).Return(user, nil)
		mockCompanyRepo.On("FindByIDOrPublicIDOrCode", ctx, comp.PublicID).Return(comp, nil)

		_, err := service.CreateUserCompany(ctx, &dto.CreateUserCompanyRequest{
			UserID:    user.PublicID,
			CompanyID: comp.PublicID,
			Role:      "super_god_role",
		})
		require.Error(t, err)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 400, de.StatusCode())
	})

	t.Run("InvalidStatus", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCompanyRepo := new(MockCompanyRepository)
		service := services.NewUserCompanyService(nil, mockUserRepo, mockCompanyRepo)

		user := &models.User{
			ID:       uuid.New(),
			PublicID: "usr_active",
			Status:   models.UserStatusActive,
		}
		comp := &models.Company{
			ID:       uuid.New(),
			PublicID: "cmp_active",
			Status:   models.CompanyStatusActive,
		}
		mockUserRepo.On("FindByIDOrPublicID", ctx, user.PublicID).Return(user, nil)
		mockCompanyRepo.On("FindByIDOrPublicIDOrCode", ctx, comp.PublicID).Return(comp, nil)

		_, err := service.CreateUserCompany(ctx, &dto.CreateUserCompanyRequest{
			UserID:    user.PublicID,
			CompanyID: comp.PublicID,
			Status:    "unknown_status",
		})
		require.Error(t, err)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 400, de.StatusCode())
	})

	t.Run("Conflict_AlreadyAssociated", func(t *testing.T) {
		mockUCRepo := new(MockUserCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		mockCompanyRepo := new(MockCompanyRepository)
		service := services.NewUserCompanyService(mockUCRepo, mockUserRepo, mockCompanyRepo)

		user := &models.User{
			ID:       uuid.New(),
			PublicID: "usr_active",
			Status:   models.UserStatusActive,
		}
		comp := &models.Company{
			ID:       uuid.New(),
			PublicID: "cmp_active",
			Status:   models.CompanyStatusActive,
		}
		existing := &models.UserCompany{
			ID:        uuid.New(),
			UserID:    user.ID,
			CompanyID: comp.ID,
			Status:    models.UserCompanyStatusActive,
		}

		mockUserRepo.On("FindByIDOrPublicID", ctx, user.PublicID).Return(user, nil)
		mockCompanyRepo.On("FindByIDOrPublicIDOrCode", ctx, comp.PublicID).Return(comp, nil)
		mockUCRepo.On("FindByUserAndCompany", ctx, user.ID, comp.ID).Return(existing, nil)

		_, err := service.CreateUserCompany(ctx, &dto.CreateUserCompanyRequest{
			UserID:    user.PublicID,
			CompanyID: comp.PublicID,
		})
		require.Error(t, err)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 409, de.StatusCode())
	})

	t.Run("ExistingCheckDBError", func(t *testing.T) {
		mockUCRepo := new(MockUserCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		mockCompanyRepo := new(MockCompanyRepository)
		service := services.NewUserCompanyService(mockUCRepo, mockUserRepo, mockCompanyRepo)

		user := &models.User{
			ID:       uuid.New(),
			PublicID: "usr_active",
			Status:   models.UserStatusActive,
		}
		comp := &models.Company{
			ID:       uuid.New(),
			PublicID: "cmp_active",
			Status:   models.CompanyStatusActive,
		}

		mockUserRepo.On("FindByIDOrPublicID", ctx, user.PublicID).Return(user, nil)
		mockCompanyRepo.On("FindByIDOrPublicIDOrCode", ctx, comp.PublicID).Return(comp, nil)
		mockUCRepo.On("FindByUserAndCompany", ctx, user.ID, comp.ID).Return(nil, errors.New("db error"))

		_, err := service.CreateUserCompany(ctx, &dto.CreateUserCompanyRequest{
			UserID:    user.PublicID,
			CompanyID: comp.PublicID,
		})
		require.Error(t, err)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 500, de.StatusCode())
	})

	t.Run("CreateRecordDBError", func(t *testing.T) {
		mockUCRepo := new(MockUserCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		mockCompanyRepo := new(MockCompanyRepository)
		service := services.NewUserCompanyService(mockUCRepo, mockUserRepo, mockCompanyRepo)

		user := &models.User{
			ID:       uuid.New(),
			PublicID: "usr_active",
			Status:   models.UserStatusActive,
		}
		comp := &models.Company{
			ID:       uuid.New(),
			PublicID: "cmp_active",
			Status:   models.CompanyStatusActive,
		}

		mockUserRepo.On("FindByIDOrPublicID", ctx, user.PublicID).Return(user, nil)
		mockCompanyRepo.On("FindByIDOrPublicIDOrCode", ctx, comp.PublicID).Return(comp, nil)
		mockUCRepo.On("FindByUserAndCompany", ctx, user.ID, comp.ID).Return(nil, gorm.ErrRecordNotFound)
		mockUCRepo.On("Create", ctx, (*gorm.DB)(nil), mock.Anything).Return(errors.New("db insert failure"))

		_, err := service.CreateUserCompany(ctx, &dto.CreateUserCompanyRequest{
			UserID:    user.PublicID,
			CompanyID: comp.PublicID,
		})
		require.Error(t, err)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 500, de.StatusCode())
	})
}

// [VN] TestUserCompanyService_GetUserCompanyByID kiểm thử tra cứu chi tiết liên kết
func TestUserCompanyService_GetUserCompanyByID(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockUCRepo := new(MockUserCompanyRepository)
		service := services.NewUserCompanyService(mockUCRepo, nil, nil)

		targetID := uuid.New()
		uc := &models.UserCompany{
			ID:        targetID,
			PublicID:  "ucmp_test123",
			Role:      "member",
			Status:    "active",
			CreatedAt: time.Now(),
		}

		mockUCRepo.On("FindByIDOrPublicID", ctx, uc.PublicID).Return(uc, nil)

		res, err := service.GetUserCompanyByID(ctx, uc.PublicID)
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, uc.PublicID, res.ID)
		assert.Equal(t, "member", res.Role)

		// [VN] Kiểm tra JSON serialization đảm bảo trường json:id chứa public_id và không để lộ UUID
		jsonBytes, err := json.Marshal(res)
		require.NoError(t, err)
		var jsonMap map[string]interface{}
		require.NoError(t, json.Unmarshal(jsonBytes, &jsonMap))
		assert.Equal(t, uc.PublicID, jsonMap["id"], "user-company response json 'id' field must contain public_id")
		assert.True(t, strings.HasPrefix(jsonMap["id"].(string), "ucmp_"), "id field must have ucmp_ prefix")
		mockUCRepo.AssertExpectations(t)
	})

	t.Run("EmptyIdentifier", func(t *testing.T) {
		service := services.NewUserCompanyService(nil, nil, nil)
		_, err := service.GetUserCompanyByID(ctx, "   ")
		require.Error(t, err)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 400, de.StatusCode())
	})

	t.Run("NotFound", func(t *testing.T) {
		mockUCRepo := new(MockUserCompanyRepository)
		service := services.NewUserCompanyService(mockUCRepo, nil, nil)

		mockUCRepo.On("FindByIDOrPublicID", ctx, "ucmp_missing").Return(nil, gorm.ErrRecordNotFound)

		_, err := service.GetUserCompanyByID(ctx, "ucmp_missing")
		require.Error(t, err)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 404, de.StatusCode())
	})

	t.Run("DBError", func(t *testing.T) {
		mockUCRepo := new(MockUserCompanyRepository)
		service := services.NewUserCompanyService(mockUCRepo, nil, nil)

		mockUCRepo.On("FindByIDOrPublicID", ctx, "ucmp_err").Return(nil, errors.New("db failure"))

		_, err := service.GetUserCompanyByID(ctx, "ucmp_err")
		require.Error(t, err)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 500, de.StatusCode())
	})
}

// [VN] TestUserCompanyService_ListUserCompanies kiểm thử danh sách liên kết có phân trang và bộ lọc
func TestUserCompanyService_ListUserCompanies(t *testing.T) {
	ctx := context.Background()

	t.Run("Success_Default", func(t *testing.T) {
		mockUCRepo := new(MockUserCompanyRepository)
		service := services.NewUserCompanyService(mockUCRepo, nil, nil)

		records := []models.UserCompany{
			{
				ID:        uuid.New(),
				PublicID:  "ucmp_1",
				Role:      "owner",
				Status:    "active",
				CreatedAt: time.Now(),
			},
			{
				ID:        uuid.New(),
				PublicID:  "ucmp_2",
				Role:      "member",
				Status:    "active",
				CreatedAt: time.Now(),
			},
		}

		mockUCRepo.On("List", ctx, "", 20, (*uuid.UUID)(nil), (*uuid.UUID)(nil), "", "", "").
			Return(records, "ucmp_next", int64(2), nil)

		res, paging, err := service.ListUserCompanies(ctx, &dto.ListUserCompaniesQuery{})
		require.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Equal(t, int64(2), paging.Total)
		assert.Equal(t, "ucmp_next", paging.NextCursor)
		mockUCRepo.AssertExpectations(t)
	})

	t.Run("Success_WithUserFilter", func(t *testing.T) {
		mockUCRepo := new(MockUserCompanyRepository)
		mockUserRepo := new(MockUserRepository)
		service := services.NewUserCompanyService(mockUCRepo, mockUserRepo, nil)

		userID := uuid.New()
		user := &models.User{
			ID:       userID,
			PublicID: "usr_123",
		}

		mockUserRepo.On("FindByIDOrPublicID", ctx, "usr_123").Return(user, nil)
		mockUCRepo.On("List", ctx, "", 20, &userID, (*uuid.UUID)(nil), "", "", "").
			Return([]models.UserCompany{}, "", int64(0), nil)

		res, paging, err := service.ListUserCompanies(ctx, &dto.ListUserCompaniesQuery{
			UserID: "usr_123",
		})
		require.NoError(t, err)
		assert.Empty(t, res)
		assert.Equal(t, int64(0), paging.Total)
		mockUserRepo.AssertExpectations(t)
		mockUCRepo.AssertExpectations(t)
	})

	t.Run("Success_WithUserNotFound_ReturnsEmpty", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		service := services.NewUserCompanyService(nil, mockUserRepo, nil)

		mockUserRepo.On("FindByIDOrPublicID", ctx, "usr_unknown").Return(nil, gorm.ErrRecordNotFound)

		res, paging, err := service.ListUserCompanies(ctx, &dto.ListUserCompaniesQuery{
			UserID: "usr_unknown",
		})
		require.NoError(t, err)
		assert.Empty(t, res)
		assert.Equal(t, int64(0), paging.Total)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Success_WithCompanyFilter", func(t *testing.T) {
		mockUCRepo := new(MockUserCompanyRepository)
		mockCompanyRepo := new(MockCompanyRepository)
		service := services.NewUserCompanyService(mockUCRepo, nil, mockCompanyRepo)

		compID := uuid.New()
		comp := &models.Company{
			ID:       compID,
			PublicID: "cmp_123",
		}

		mockCompanyRepo.On("FindByIDOrPublicIDOrCode", ctx, "cmp_123").Return(comp, nil)
		mockUCRepo.On("List", ctx, "", 20, (*uuid.UUID)(nil), &compID, "", "", "").
			Return([]models.UserCompany{}, "", int64(0), nil)

		res, paging, err := service.ListUserCompanies(ctx, &dto.ListUserCompaniesQuery{
			CompanyID: "cmp_123",
		})
		require.NoError(t, err)
		assert.Empty(t, res)
		assert.Equal(t, int64(0), paging.Total)
		mockCompanyRepo.AssertExpectations(t)
		mockUCRepo.AssertExpectations(t)
	})

	t.Run("Success_WithCompanyNotFound_ReturnsEmpty", func(t *testing.T) {
		mockCompanyRepo := new(MockCompanyRepository)
		service := services.NewUserCompanyService(nil, nil, mockCompanyRepo)

		mockCompanyRepo.On("FindByIDOrPublicIDOrCode", ctx, "cmp_unknown").Return(nil, gorm.ErrRecordNotFound)

		res, paging, err := service.ListUserCompanies(ctx, &dto.ListUserCompaniesQuery{
			CompanyID: "cmp_unknown",
		})
		require.NoError(t, err)
		assert.Empty(t, res)
		assert.Equal(t, int64(0), paging.Total)
		mockCompanyRepo.AssertExpectations(t)
	})

	t.Run("DBError", func(t *testing.T) {
		mockUCRepo := new(MockUserCompanyRepository)
		service := services.NewUserCompanyService(mockUCRepo, nil, nil)

		mockUCRepo.On("List", ctx, "", 20, (*uuid.UUID)(nil), (*uuid.UUID)(nil), "", "", "").
			Return(nil, "", int64(0), errors.New("list query failed"))

		_, _, err := service.ListUserCompanies(ctx, &dto.ListUserCompaniesQuery{})
		require.Error(t, err)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 500, de.StatusCode())
	})
}

// [VN] TestUserCompanyService_UpdateUserCompany kiểm thử cập nhật liên kết
func TestUserCompanyService_UpdateUserCompany(t *testing.T) {
	ctx := context.Background()

	t.Run("Success_UpdateFields", func(t *testing.T) {
		mockUCRepo := new(MockUserCompanyRepository)
		service := services.NewUserCompanyService(mockUCRepo, nil, nil)

		uc := &models.UserCompany{
			ID:        uuid.New(),
			PublicID:  "ucmp_update",
			Role:      "member",
			Status:    "active",
			CreatedAt: time.Now(),
		}

		newRole := "admin"
		newTitle := "Lead Architect"
		newDept := "R&D"
		newStatus := "inactive"
		req := &dto.UpdateUserCompanyRequest{
			Role:       &newRole,
			JobTitle:   &newTitle,
			Department: &newDept,
			Status:     &newStatus,
			Permissions: map[string]interface{}{
				"manage_all": true,
			},
			Metadata: map[string]interface{}{
				"tag": "core",
			},
		}

		mockUCRepo.On("FindByIDOrPublicID", ctx, uc.PublicID).Return(uc, nil)
		mockUCRepo.On("Update", ctx, (*gorm.DB)(nil), mock.MatchedBy(func(u *models.UserCompany) bool {
			return u.Role == "admin" && *u.JobTitle == "Lead Architect" && u.Status == "inactive"
		})).Return(nil)

		res, err := service.UpdateUserCompany(ctx, uc.PublicID, req)
		require.NoError(t, err)
		assert.Equal(t, "admin", res.Role)
		assert.Equal(t, &newTitle, res.JobTitle)
		assert.Equal(t, &newDept, res.Department)
		assert.Equal(t, "inactive", res.Status)
		mockUCRepo.AssertExpectations(t)
	})

	t.Run("EmptyIdentifier", func(t *testing.T) {
		service := services.NewUserCompanyService(nil, nil, nil)
		_, err := service.UpdateUserCompany(ctx, "", &dto.UpdateUserCompanyRequest{})
		require.Error(t, err)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 400, de.StatusCode())
	})

	t.Run("NotFound", func(t *testing.T) {
		mockUCRepo := new(MockUserCompanyRepository)
		service := services.NewUserCompanyService(mockUCRepo, nil, nil)

		mockUCRepo.On("FindByIDOrPublicID", ctx, "ucmp_missing").Return(nil, gorm.ErrRecordNotFound)

		_, err := service.UpdateUserCompany(ctx, "ucmp_missing", &dto.UpdateUserCompanyRequest{})
		require.Error(t, err)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 404, de.StatusCode())
	})

	t.Run("InvalidRole", func(t *testing.T) {
		mockUCRepo := new(MockUserCompanyRepository)
		service := services.NewUserCompanyService(mockUCRepo, nil, nil)

		uc := &models.UserCompany{
			ID:       uuid.New(),
			PublicID: "ucmp_test",
		}
		mockUCRepo.On("FindByIDOrPublicID", ctx, uc.PublicID).Return(uc, nil)

		invalidRole := "invalid_king"
		_, err := service.UpdateUserCompany(ctx, uc.PublicID, &dto.UpdateUserCompanyRequest{
			Role: &invalidRole,
		})
		require.Error(t, err)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 400, de.StatusCode())
	})

	t.Run("InvalidStatus", func(t *testing.T) {
		mockUCRepo := new(MockUserCompanyRepository)
		service := services.NewUserCompanyService(mockUCRepo, nil, nil)

		uc := &models.UserCompany{
			ID:       uuid.New(),
			PublicID: "ucmp_test",
		}
		mockUCRepo.On("FindByIDOrPublicID", ctx, uc.PublicID).Return(uc, nil)

		invalidStatus := "not_real_status"
		_, err := service.UpdateUserCompany(ctx, uc.PublicID, &dto.UpdateUserCompanyRequest{
			Status: &invalidStatus,
		})
		require.Error(t, err)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 400, de.StatusCode())
	})

	t.Run("UpdateDBError", func(t *testing.T) {
		mockUCRepo := new(MockUserCompanyRepository)
		service := services.NewUserCompanyService(mockUCRepo, nil, nil)

		uc := &models.UserCompany{
			ID:       uuid.New(),
			PublicID: "ucmp_test",
		}
		mockUCRepo.On("FindByIDOrPublicID", ctx, uc.PublicID).Return(uc, nil)
		mockUCRepo.On("Update", ctx, (*gorm.DB)(nil), mock.Anything).Return(errors.New("db error"))

		role := "admin"
		_, err := service.UpdateUserCompany(ctx, uc.PublicID, &dto.UpdateUserCompanyRequest{
			Role: &role,
		})
		require.Error(t, err)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 500, de.StatusCode())
	})
}

// [VN] TestUserCompanyService_DeleteUserCompany kiểm thử xóa mềm và xóa cứng liên kết
func TestUserCompanyService_DeleteUserCompany(t *testing.T) {
	ctx := context.Background()

	t.Run("Success_SoftDelete", func(t *testing.T) {
		mockUCRepo := new(MockUserCompanyRepository)
		service := services.NewUserCompanyService(mockUCRepo, nil, nil)

		targetID := uuid.New()
		uc := &models.UserCompany{
			ID:       targetID,
			PublicID: "ucmp_del1",
		}

		mockUCRepo.On("FindByIDOrPublicID", ctx, uc.PublicID).Return(uc, nil)
		mockUCRepo.On("Delete", ctx, (*gorm.DB)(nil), targetID, true).Return(nil)

		err := service.DeleteUserCompany(ctx, uc.PublicID, false)
		require.NoError(t, err)
		mockUCRepo.AssertExpectations(t)
	})

	t.Run("Success_HardDelete", func(t *testing.T) {
		mockUCRepo := new(MockUserCompanyRepository)
		service := services.NewUserCompanyService(mockUCRepo, nil, nil)

		targetID := uuid.New()
		uc := &models.UserCompany{
			ID:       targetID,
			PublicID: "ucmp_del2",
		}

		mockUCRepo.On("FindByIDOrPublicID", ctx, uc.PublicID).Return(uc, nil)
		mockUCRepo.On("Delete", ctx, (*gorm.DB)(nil), targetID, false).Return(nil)

		err := service.DeleteUserCompany(ctx, uc.PublicID, true)
		require.NoError(t, err)
		mockUCRepo.AssertExpectations(t)
	})

	t.Run("EmptyIdentifier", func(t *testing.T) {
		service := services.NewUserCompanyService(nil, nil, nil)
		err := service.DeleteUserCompany(ctx, "", false)
		require.Error(t, err)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 400, de.StatusCode())
	})

	t.Run("NotFound", func(t *testing.T) {
		mockUCRepo := new(MockUserCompanyRepository)
		service := services.NewUserCompanyService(mockUCRepo, nil, nil)

		mockUCRepo.On("FindByIDOrPublicID", ctx, "ucmp_missing").Return(nil, gorm.ErrRecordNotFound)

		err := service.DeleteUserCompany(ctx, "ucmp_missing", false)
		require.Error(t, err)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 404, de.StatusCode())
	})

	t.Run("DBError", func(t *testing.T) {
		mockUCRepo := new(MockUserCompanyRepository)
		service := services.NewUserCompanyService(mockUCRepo, nil, nil)

		targetID := uuid.New()
		uc := &models.UserCompany{
			ID:       targetID,
			PublicID: "ucmp_err",
		}

		mockUCRepo.On("FindByIDOrPublicID", ctx, uc.PublicID).Return(uc, nil)
		mockUCRepo.On("Delete", ctx, (*gorm.DB)(nil), targetID, true).Return(errors.New("db delete error"))

		err := service.DeleteUserCompany(ctx, uc.PublicID, false)
		require.Error(t, err)
		var de *common.DefaultError
		require.True(t, errors.As(err, &de))
		assert.Equal(t, 500, de.StatusCode())
	})
}
