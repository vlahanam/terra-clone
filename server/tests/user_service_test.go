package tests

import (
	"context"
	"errors"
	"testing"

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

// [VN] TestUserService_CreateUser kiểm thử nghiệp vụ tạo người dùng mới kèm mật khẩu và hồ sơ
func TestUserService_CreateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)
		gormDB, sqlMock := newMockGormDB(t)

		service := services.NewUserService(mockUserRepo, mockCredRepo, mockProfileRepo)

		phone := "0912345678"
		dob := "1995-05-20"
		gender := "male"
		address := "123 Terra Street"
		countryCode := "VN"
		nationality := "Vietnamese"
		timezone := "Asia/Ho_Chi_Minh"
		locale := "vi-VN"
		city := "Hanoi"
		state := "Hanoi"
		postalCode := "100000"

		req := &dto.CreateUserRequest{
			Email:           "create_user@example.com",
			Password:        "SecureAdminPass123!",
			FullName:        "Created User",
			PhoneNumber:     &phone,
			DateOfBirth:     &dob,
			Gender:          &gender,
			Address:         &address,
			CountryCode:     &countryCode,
			Nationality:     &nationality,
			Timezone:        &timezone,
			Locale:          &locale,
			City:            &city,
			StateOrProvince: &state,
			PostalCode:      &postalCode,
			Status:          models.UserStatusActive,
		}

		mockUserRepo.On("FindByEmail", ctx, req.Email).Return(nil, gorm.ErrRecordNotFound)
		mockUserRepo.On("GetDB").Return(gormDB)

		sqlMock.ExpectBegin()
		mockUserRepo.On("Create", ctx, mock.AnythingOfType("*gorm.DB"), mock.MatchedBy(func(u *models.User) bool {
			return u.Email == req.Email && u.FullName == req.FullName
		})).Return(nil)

		mockCredRepo.On("Create", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*models.UserCredentials")).Return(nil)
		mockProfileRepo.On("Create", ctx, mock.AnythingOfType("*gorm.DB"), mock.MatchedBy(func(p *models.UserProfile) bool {
			return p.CountryCode != nil && *p.CountryCode == "VN" && p.Nationality != nil && *p.Nationality == "Vietnamese"
		})).Return(nil)
		sqlMock.ExpectCommit()

		resp, err := service.CreateUser(ctx, req)
		require.NoError(t, err, "user creation should succeed")
		require.NotNil(t, resp, "response should not be nil")
		assert.Equal(t, req.Email, resp.Email, "email should match")
		assert.Equal(t, req.FullName, resp.FullName, "fullname should match")
		assert.Equal(t, models.UserStatusActive, resp.Status, "status should be active")
		require.NotNil(t, resp.Profile, "profile response should not be nil")
		assert.Equal(t, req.PhoneNumber, resp.Profile.PhoneNumber, "phone number should match")
		assert.Equal(t, &countryCode, resp.Profile.CountryCode, "country code should match")
		assert.Equal(t, &nationality, resp.Profile.Nationality, "nationality should match")
		assert.Equal(t, &timezone, resp.Profile.Timezone, "timezone should match")
		assert.Equal(t, &city, resp.Profile.City, "city should match")

		mockUserRepo.AssertExpectations(t)
		mockCredRepo.AssertExpectations(t)
		mockProfileRepo.AssertExpectations(t)
		assert.NoError(t, sqlMock.ExpectationsWereMet(), "sql expectations met")
	})

	t.Run("EmailAlreadyExists", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewUserService(mockUserRepo, mockCredRepo, mockProfileRepo)

		req := &dto.CreateUserRequest{
			Email:    "dup@example.com",
			Password: "Password123!",
			FullName: "Duplicate Email User",
		}

		mockUserRepo.On("FindByEmail", ctx, req.Email).Return(&models.User{
			ID:    uuid.New(),
			Email: req.Email,
		}, nil)

		resp, err := service.CreateUser(ctx, req)
		require.Error(t, err, "should return conflict error")
		assert.Nil(t, resp, "response should be nil")

		var appErr *common.DefaultError
		ok := errors.As(err, &appErr)
		require.True(t, ok, "error should be common.DefaultError")
		assert.Equal(t, 409, appErr.StatusCode(), "status should be 409 Conflict")

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("TransactionFailure", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)
		gormDB, sqlMock := newMockGormDB(t)

		service := services.NewUserService(mockUserRepo, mockCredRepo, mockProfileRepo)

		req := &dto.CreateUserRequest{
			Email:    "txfail@example.com",
			Password: "Password123!",
			FullName: "Tx Fail",
		}

		mockUserRepo.On("FindByEmail", ctx, req.Email).Return(nil, gorm.ErrRecordNotFound)
		mockUserRepo.On("GetDB").Return(gormDB)

		sqlMock.ExpectBegin()
		mockUserRepo.On("Create", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*models.User")).
			Return(errors.New("db error on insert user"))
		sqlMock.ExpectRollback()

		resp, err := service.CreateUser(ctx, req)
		require.Error(t, err, "should return error on tx failure")
		assert.Nil(t, resp, "response should be nil")

		var appErr *common.DefaultError
		ok := errors.As(err, &appErr)
		require.True(t, ok, "error should be common.DefaultError")
		assert.Equal(t, 500, appErr.StatusCode(), "status should be 500")

		mockUserRepo.AssertExpectations(t)
		assert.NoError(t, sqlMock.ExpectationsWereMet(), "sql expectations met")
	})
}

// [VN] TestUserService_GetUserByID kiểm thử tra cứu người dùng qua UUID hoặc PublicID
func TestUserService_GetUserByID(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	publicID := "usr_lookup123"

	t.Run("Success", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewUserService(mockUserRepo, mockCredRepo, mockProfileRepo)

		user := &models.User{
			ID:       userID,
			PublicID: publicID,
			Email:    "lookup@example.com",
			FullName: "Lookup User",
			Status:   models.UserStatusActive,
		}
		mockUserRepo.On("FindByIDOrPublicID", ctx, publicID).Return(user, nil)

		resp, err := service.GetUserByID(ctx, publicID)
		require.NoError(t, err, "get user by id should succeed")
		require.NotNil(t, resp, "response should not be nil")
		assert.Equal(t, user.Email, resp.Email, "email should match")
		assert.Equal(t, user.PublicID, resp.PublicID, "public_id should match")

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewUserService(mockUserRepo, mockCredRepo, mockProfileRepo)

		mockUserRepo.On("FindByIDOrPublicID", ctx, "unknown_id").Return(nil, gorm.ErrRecordNotFound)

		resp, err := service.GetUserByID(ctx, "unknown_id")
		require.Error(t, err, "should return error when user not found")
		assert.Nil(t, resp, "response should be nil")

		var appErr *common.DefaultError
		ok := errors.As(err, &appErr)
		require.True(t, ok, "error should be common.DefaultError")
		assert.Equal(t, 404, appErr.StatusCode(), "status should be 404 Not Found")

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewUserService(mockUserRepo, mockCredRepo, mockProfileRepo)

		mockUserRepo.On("FindByIDOrPublicID", ctx, publicID).Return(nil, errors.New("db disconnect"))

		resp, err := service.GetUserByID(ctx, publicID)
		require.Error(t, err, "should return internal error on db error")
		assert.Nil(t, resp, "response should be nil")

		var appErr *common.DefaultError
		ok := errors.As(err, &appErr)
		require.True(t, ok, "error should be common.DefaultError")
		assert.Equal(t, 500, appErr.StatusCode(), "status should be 500")

		mockUserRepo.AssertExpectations(t)
	})
}

// [VN] TestUserService_ListUsers kiểm thử truy vấn danh sách người dùng phân trang con trỏ
func TestUserService_ListUsers(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewUserService(mockUserRepo, mockCredRepo, mockProfileRepo)

		query := &dto.ListUsersQuery{
			Cursor: "018f2a1b-9f3b-7f28-8d54-1b945f3408f1",
			Limit:  10,
			Search: "John",
			Status: "active",
		}

		mockUsers := []models.User{
			{
				ID:       uuid.New(),
				PublicID: "usr_1",
				Email:    "john1@example.com",
				FullName: "John Doe",
				Status:   models.UserStatusActive,
			},
			{
				ID:       uuid.New(),
				PublicID: "usr_2",
				Email:    "john2@example.com",
				FullName: "John Smith",
				Status:   models.UserStatusActive,
			},
		}
		nextCursor := "018f2a1c-4b12-7e90-a321-2e841a1239bc"
		total := int64(25)

		mockUserRepo.On("List", ctx, query.Cursor, query.Limit, query.Search, query.Status, "").
			Return(mockUsers, nextCursor, total, nil)

		users, paging, err := service.ListUsers(ctx, query)
		require.NoError(t, err, "list users should succeed")
		assert.Len(t, users, 2, "should return 2 users")
		require.NotNil(t, paging, "paging should not be nil")
		assert.Equal(t, total, paging.Total, "total should match")
		assert.Equal(t, nextCursor, paging.NextCursor, "next cursor should match")
		assert.Equal(t, query.Cursor, paging.FakeCursor, "cursor should match fake_cursor")

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Success_WithCountryFilter", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewUserService(mockUserRepo, mockCredRepo, mockProfileRepo)

		query := &dto.ListUsersQuery{
			CountryCode: "VN",
			Limit:       10,
		}

		vnUser := models.User{
			ID:       uuid.New(),
			PublicID: "usr_vn",
			Email:    "vn@example.com",
			FullName: "VN User",
			Profile: &models.UserProfile{
				CountryCode: func(s string) *string { return &s }("VN"),
			},
		}

		mockUserRepo.On("List", ctx, "", 10, "", "", "VN").
			Return([]models.User{vnUser}, "", int64(1), nil)

		users, paging, err := service.ListUsers(ctx, query)
		require.NoError(t, err)
		require.Len(t, users, 1)
		assert.Equal(t, "usr_vn", users[0].PublicID)
		assert.Equal(t, int64(1), paging.Total)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewUserService(mockUserRepo, mockCredRepo, mockProfileRepo)

		query := &dto.ListUsersQuery{Limit: 20}
		mockUserRepo.On("List", ctx, query.Cursor, query.Limit, query.Search, query.Status, "").
			Return(nil, "", int64(0), errors.New("db error"))

		users, paging, err := service.ListUsers(ctx, query)
		require.Error(t, err, "should return error on db failure")
		assert.Nil(t, users, "users should be nil")
		assert.Nil(t, paging, "paging should be nil")

		var appErr *common.DefaultError
		ok := errors.As(err, &appErr)
		require.True(t, ok, "error should be common.DefaultError")
		assert.Equal(t, 500, appErr.StatusCode(), "status should be 500")

		mockUserRepo.AssertExpectations(t)
	})
}

// [VN] TestUserService_UpdateUser kiểm thử cập nhật thông tin tài khoản và upsert hồ sơ
func TestUserService_UpdateUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	publicID := "usr_update123"

	t.Run("Success_UpdateBasicAndProfile", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)
		gormDB, sqlMock := newMockGormDB(t)

		service := services.NewUserService(mockUserRepo, mockCredRepo, mockProfileRepo)

		existingUser := &models.User{
			ID:       userID,
			PublicID: publicID,
			Email:    "update@example.com",
			FullName: "Old Name",
			Status:   models.UserStatusActive,
			Profile: &models.UserProfile{
				UserID:      userID,
				PhoneNumber: nil,
			},
		}

		newFullName := "Updated Name"
		newPhone := "0987654321"
		newStatus := models.UserStatusInactive
		newCountry := "SG"
		newCity := "Singapore"
		req := &dto.UpdateUserRequest{
			FullName:    &newFullName,
			PhoneNumber: &newPhone,
			CountryCode: &newCountry,
			City:        &newCity,
			Status:      &newStatus,
		}

		mockUserRepo.On("FindByIDOrPublicID", ctx, publicID).Return(existingUser, nil)
		mockUserRepo.On("GetDB").Return(gormDB)

		sqlMock.ExpectBegin()
		mockUserRepo.On("Update", ctx, mock.AnythingOfType("*gorm.DB"), mock.MatchedBy(func(u *models.User) bool {
			return u.FullName == newFullName && u.Status == newStatus
		})).Return(nil)

		mockProfileRepo.On("Upsert", ctx, mock.AnythingOfType("*gorm.DB"), mock.MatchedBy(func(p *models.UserProfile) bool {
			return p.PhoneNumber != nil && *p.PhoneNumber == newPhone &&
				p.CountryCode != nil && *p.CountryCode == newCountry &&
				p.City != nil && *p.City == newCity
		})).Return(nil)
		sqlMock.ExpectCommit()

		resp, err := service.UpdateUser(ctx, publicID, req)
		require.NoError(t, err, "update user should succeed")
		require.NotNil(t, resp, "response should not be nil")
		assert.Equal(t, newFullName, resp.FullName, "full name should be updated")
		assert.Equal(t, newStatus, resp.Status, "status should be updated")
		require.NotNil(t, resp.Profile, "profile should not be nil")
		assert.Equal(t, &newCountry, resp.Profile.CountryCode, "country code should be updated")
		assert.Equal(t, &newCity, resp.Profile.City, "city should be updated")

		mockUserRepo.AssertExpectations(t)
		mockProfileRepo.AssertExpectations(t)
		assert.NoError(t, sqlMock.ExpectationsWereMet(), "sql expectations met")
	})

	t.Run("NotFound", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewUserService(mockUserRepo, mockCredRepo, mockProfileRepo)

		mockUserRepo.On("FindByIDOrPublicID", ctx, "notfound").Return(nil, gorm.ErrRecordNotFound)

		newName := "New Name"
		resp, err := service.UpdateUser(ctx, "notfound", &dto.UpdateUserRequest{FullName: &newName})
		require.Error(t, err, "should return not found error")
		assert.Nil(t, resp, "response should be nil")

		var appErr *common.DefaultError
		ok := errors.As(err, &appErr)
		require.True(t, ok, "error should be common.DefaultError")
		assert.Equal(t, 404, appErr.StatusCode(), "status should be 404")

		mockUserRepo.AssertExpectations(t)
	})
}

// [VN] TestUserService_DeleteUser kiểm thử xóa mềm và xóa cứng người dùng
func TestUserService_DeleteUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	publicID := "usr_del123"

	existingUser := &models.User{
		ID:       userID,
		PublicID: publicID,
		Email:    "del@example.com",
		Status:   models.UserStatusActive,
	}

	t.Run("SoftDelete_Success", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewUserService(mockUserRepo, mockCredRepo, mockProfileRepo)

		mockUserRepo.On("FindByIDOrPublicID", ctx, publicID).Return(existingUser, nil)
		mockUserRepo.On("Delete", ctx, (*gorm.DB)(nil), userID, true).Return(nil)

		err := service.DeleteUser(ctx, publicID, false)
		require.NoError(t, err, "soft delete should succeed")

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("HardDelete_Success", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewUserService(mockUserRepo, mockCredRepo, mockProfileRepo)

		mockUserRepo.On("FindByIDOrPublicID", ctx, publicID).Return(existingUser, nil)
		mockUserRepo.On("Delete", ctx, (*gorm.DB)(nil), userID, false).Return(nil)

		err := service.DeleteUser(ctx, publicID, true)
		require.NoError(t, err, "hard delete should succeed")

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewUserService(mockUserRepo, mockCredRepo, mockProfileRepo)

		mockUserRepo.On("FindByIDOrPublicID", ctx, "nonexistent").Return(nil, gorm.ErrRecordNotFound)

		err := service.DeleteUser(ctx, "nonexistent", false)
		require.Error(t, err, "should return not found error")

		var appErr *common.DefaultError
		ok := errors.As(err, &appErr)
		require.True(t, ok, "error should be common.DefaultError")
		assert.Equal(t, 404, appErr.StatusCode(), "status should be 404")

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewUserService(mockUserRepo, mockCredRepo, mockProfileRepo)

		mockUserRepo.On("FindByIDOrPublicID", ctx, publicID).Return(existingUser, nil)
		mockUserRepo.On("Delete", ctx, (*gorm.DB)(nil), userID, true).Return(errors.New("db delete error"))

		err := service.DeleteUser(ctx, publicID, false)
		require.Error(t, err, "should return error on db delete failure")

		var appErr *common.DefaultError
		ok := errors.As(err, &appErr)
		require.True(t, ok, "error should be common.DefaultError")
		assert.Equal(t, 500, appErr.StatusCode(), "status should be 500")

		mockUserRepo.AssertExpectations(t)
	})
}
