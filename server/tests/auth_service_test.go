package tests

import (
	"context"
	"errors"
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
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const testJWTSecret = "test-jwt-secret-key-super-secure-32bytes!"

// [VN] TestAuthService_Register kiểm thử toàn diện nghiệp vụ đăng ký tài khoản mới của AuthService
func TestAuthService_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)
		gormDB, sqlMock := newMockGormDB(t)

		service := services.NewAuthService(mockUserRepo, mockCredRepo, mockProfileRepo, testJWTSecret)

		req := &dto.RegisterRequest{
			Email:       "newuser@example.com",
			Password:    "SecurePass123!",
			FullName:    "Terra User",
			PhoneNumber: "0901234567",
		}

		mockUserRepo.On("FindByEmail", ctx, req.Email).Return(nil, gorm.ErrRecordNotFound)
		mockUserRepo.On("GetDB").Return(gormDB)

		sqlMock.ExpectBegin()
		mockUserRepo.On("Create", ctx, mock.AnythingOfType("*gorm.DB"), mock.MatchedBy(func(u *models.User) bool {
			return u.Email == req.Email && u.FullName == req.FullName
		})).Return(nil)

		mockCredRepo.On("Create", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*models.UserCredentials")).Return(nil)
		mockProfileRepo.On("Create", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*models.UserProfile")).Return(nil)
		sqlMock.ExpectCommit()

		resp, err := service.Register(ctx, req)
		require.NoError(t, err, "register should succeed without errors")
		require.NotNil(t, resp, "response should not be nil")
		assert.NotEmpty(t, resp.AccessToken, "access token should not be empty")
		assert.NotEmpty(t, resp.RefreshToken, "refresh token should not be empty")
		assert.Equal(t, "Bearer", resp.TokenType, "token type should be Bearer")
		assert.Equal(t, req.Email, resp.User.Email, "email in response should match request")

		mockUserRepo.AssertExpectations(t)
		mockCredRepo.AssertExpectations(t)
		mockProfileRepo.AssertExpectations(t)
		assert.NoError(t, sqlMock.ExpectationsWereMet(), "sqlmock expectations must be fulfilled")
	})

	t.Run("EmailAlreadyExists", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewAuthService(mockUserRepo, mockCredRepo, mockProfileRepo, testJWTSecret)

		req := &dto.RegisterRequest{
			Email:    "existing@example.com",
			Password: "SecurePass123!",
			FullName: "Existing User",
		}

		existingUser := &models.User{
			ID:    uuid.New(),
			Email: req.Email,
		}
		mockUserRepo.On("FindByEmail", ctx, req.Email).Return(existingUser, nil)

		resp, err := service.Register(ctx, req)
		require.Error(t, err, "should return error when email already exists")
		assert.Nil(t, resp, "response should be nil on error")

		var appErr *common.DefaultError
		ok := errors.As(err, &appErr)
		require.True(t, ok, "error should be an instance of common.DefaultError")
		assert.Equal(t, 409, appErr.StatusCode(), "status code should be 409 Conflict")

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("DatabaseQueryError", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewAuthService(mockUserRepo, mockCredRepo, mockProfileRepo, testJWTSecret)

		req := &dto.RegisterRequest{
			Email:    "queryerror@example.com",
			Password: "SecurePass123!",
		}

		mockUserRepo.On("FindByEmail", ctx, req.Email).Return(nil, errors.New("connection reset by peer"))

		resp, err := service.Register(ctx, req)
		require.Error(t, err, "should return internal error when db query fails")
		assert.Nil(t, resp, "response should be nil")

		var appErr *common.DefaultError
		ok := errors.As(err, &appErr)
		require.True(t, ok, "error should be common.DefaultError")
		assert.Equal(t, 500, appErr.StatusCode(), "status code should be 500 Internal Server Error")

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("TransactionFailure", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)
		gormDB, sqlMock := newMockGormDB(t)

		service := services.NewAuthService(mockUserRepo, mockCredRepo, mockProfileRepo, testJWTSecret)

		req := &dto.RegisterRequest{
			Email:    "txfail@example.com",
			Password: "SecurePass123!",
		}

		mockUserRepo.On("FindByEmail", ctx, req.Email).Return(nil, gorm.ErrRecordNotFound)
		mockUserRepo.On("GetDB").Return(gormDB)

		sqlMock.ExpectBegin()
		mockUserRepo.On("Create", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*models.User")).
			Return(errors.New("insert into users error"))
		sqlMock.ExpectRollback()

		resp, err := service.Register(ctx, req)
		require.Error(t, err, "should return error when transaction fails")
		assert.Nil(t, resp, "response should be nil")

		var appErr *common.DefaultError
		ok := errors.As(err, &appErr)
		require.True(t, ok, "error should be common.DefaultError")
		assert.Equal(t, 500, appErr.StatusCode(), "status code should be 500")

		mockUserRepo.AssertExpectations(t)
		assert.NoError(t, sqlMock.ExpectationsWereMet(), "sql expectations should be met")
	})
}

// [VN] TestAuthService_Login kiểm thử nghiệp vụ đăng nhập, brute-force lockout và so khớp mật khẩu
func TestAuthService_Login(t *testing.T) {
	ctx := context.Background()
	plainPassword := "CorrectPassword123!"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	require.NoError(t, err, "failed to hash password for testing")

	userID := uuid.New()
	baseUser := &models.User{
		ID:       userID,
		PublicID: "usr_test123",
		Email:    "loginuser@example.com",
		Status:   models.UserStatusActive,
		Credentials: &models.UserCredentials{
			UserID:               userID,
			PasswordHash:         string(hashedPassword),
			TokenVersion:         1,
			FailedLoginAttempts:  0,
			LockedUntil:          nil,
		},
	}

	t.Run("Success", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewAuthService(mockUserRepo, mockCredRepo, mockProfileRepo, testJWTSecret)

		mockUserRepo.On("FindByEmail", ctx, baseUser.Email).Return(baseUser, nil)
		mockCredRepo.On("ResetFailedLogin", ctx, (*gorm.DB)(nil), baseUser.ID).Return(nil)

		resp, err := service.Login(ctx, &dto.LoginRequest{
			Email:    baseUser.Email,
			Password: plainPassword,
		})

		require.NoError(t, err, "login should succeed")
		require.NotNil(t, resp, "response should not be nil")
		assert.NotEmpty(t, resp.AccessToken, "access token should not be empty")
		assert.Equal(t, baseUser.Email, resp.User.Email, "email should match")

		mockUserRepo.AssertExpectations(t)
		mockCredRepo.AssertExpectations(t)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewAuthService(mockUserRepo, mockCredRepo, mockProfileRepo, testJWTSecret)

		mockUserRepo.On("FindByEmail", ctx, "notfound@example.com").Return(nil, gorm.ErrRecordNotFound)

		resp, err := service.Login(ctx, &dto.LoginRequest{
			Email:    "notfound@example.com",
			Password: "Password123!",
		})

		require.Error(t, err, "should return unauthorized")
		assert.Nil(t, resp, "response should be nil")

		var appErr *common.DefaultError
		ok := errors.As(err, &appErr)
		require.True(t, ok, "error should be common.DefaultError")
		assert.Equal(t, 401, appErr.StatusCode(), "status should be 401 Unauthorized")

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("UserDeleted", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewAuthService(mockUserRepo, mockCredRepo, mockProfileRepo, testJWTSecret)

		deletedUser := &models.User{
			ID:     uuid.New(),
			Email:  "deleted@example.com",
			Status: models.UserStatusDeleted,
		}
		mockUserRepo.On("FindByEmail", ctx, deletedUser.Email).Return(deletedUser, nil)

		resp, err := service.Login(ctx, &dto.LoginRequest{
			Email:    deletedUser.Email,
			Password: plainPassword,
		})

		require.Error(t, err, "should return error for deleted user")
		assert.Nil(t, resp, "response should be nil")

		var appErr *common.DefaultError
		ok := errors.As(err, &appErr)
		require.True(t, ok, "error should be common.DefaultError")
		assert.Equal(t, 401, appErr.StatusCode(), "status should be 401")

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("UserInactive", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewAuthService(mockUserRepo, mockCredRepo, mockProfileRepo, testJWTSecret)

		inactiveUser := &models.User{
			ID:     uuid.New(),
			Email:  "inactive@example.com",
			Status: models.UserStatusInactive,
		}
		mockUserRepo.On("FindByEmail", ctx, inactiveUser.Email).Return(inactiveUser, nil)

		resp, err := service.Login(ctx, &dto.LoginRequest{
			Email:    inactiveUser.Email,
			Password: plainPassword,
		})

		require.Error(t, err, "should return forbidden for inactive user")
		assert.Nil(t, resp, "response should be nil")

		var appErr *common.DefaultError
		ok := errors.As(err, &appErr)
		require.True(t, ok, "error should be common.DefaultError")
		assert.Equal(t, 403, appErr.StatusCode(), "status should be 403 Forbidden")

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("AccountTemporarilyLocked", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewAuthService(mockUserRepo, mockCredRepo, mockProfileRepo, testJWTSecret)

		futureLock := time.Now().Add(10 * time.Minute)
		lockedUser := &models.User{
			ID:     uuid.New(),
			Email:  "locked@example.com",
			Status: models.UserStatusActive,
			Credentials: &models.UserCredentials{
				PasswordHash: string(hashedPassword),
				LockedUntil:  &futureLock,
			},
		}
		mockUserRepo.On("FindByEmail", ctx, lockedUser.Email).Return(lockedUser, nil)

		resp, err := service.Login(ctx, &dto.LoginRequest{
			Email:    lockedUser.Email,
			Password: plainPassword,
		})

		require.Error(t, err, "should reject login for locked account")
		assert.Nil(t, resp, "response should be nil")

		var appErr *common.DefaultError
		ok := errors.As(err, &appErr)
		require.True(t, ok, "error should be common.DefaultError")
		assert.Equal(t, 403, appErr.StatusCode(), "status should be 403 Forbidden")

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("WrongPassword_RemainingAttempts", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewAuthService(mockUserRepo, mockCredRepo, mockProfileRepo, testJWTSecret)

		mockUserRepo.On("FindByEmail", ctx, baseUser.Email).Return(baseUser, nil)

		updatedCreds := &models.UserCredentials{
			UserID:              baseUser.ID,
			FailedLoginAttempts: 2,
			LockedUntil:         nil,
		}
		mockCredRepo.On("RecordFailedLogin", ctx, (*gorm.DB)(nil), baseUser.ID, common.ACC_MAX_LOGIN_ATTEMPTS, common.ACC_LOCK_TIME).
			Return(updatedCreds, nil)

		resp, err := service.Login(ctx, &dto.LoginRequest{
			Email:    baseUser.Email,
			Password: "WrongPassword123!",
		})

		require.Error(t, err, "should return unauthorized for wrong password")
		assert.Nil(t, resp, "response should be nil")

		var appErr *common.DefaultError
		ok := errors.As(err, &appErr)
		require.True(t, ok, "error should be common.DefaultError")
		assert.Equal(t, 401, appErr.StatusCode(), "status should be 401 Unauthorized")

		mockUserRepo.AssertExpectations(t)
		mockCredRepo.AssertExpectations(t)
	})

	t.Run("WrongPassword_ExceedsMaxAttemptsAndLocks", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewAuthService(mockUserRepo, mockCredRepo, mockProfileRepo, testJWTSecret)

		mockUserRepo.On("FindByEmail", ctx, baseUser.Email).Return(baseUser, nil)

		lockedUntil := time.Now().Add(15 * time.Minute)
		updatedCreds := &models.UserCredentials{
			UserID:              baseUser.ID,
			FailedLoginAttempts: 5,
			LockedUntil:         &lockedUntil,
		}
		mockCredRepo.On("RecordFailedLogin", ctx, (*gorm.DB)(nil), baseUser.ID, common.ACC_MAX_LOGIN_ATTEMPTS, common.ACC_LOCK_TIME).
			Return(updatedCreds, nil)

		resp, err := service.Login(ctx, &dto.LoginRequest{
			Email:    baseUser.Email,
			Password: "WrongPassword123!",
		})

		require.Error(t, err, "should return forbidden when account gets locked")
		assert.Nil(t, resp, "response should be nil")

		var appErr *common.DefaultError
		ok := errors.As(err, &appErr)
		require.True(t, ok, "error should be common.DefaultError")
		assert.Equal(t, 403, appErr.StatusCode(), "status should be 403 Forbidden")

		mockUserRepo.AssertExpectations(t)
		mockCredRepo.AssertExpectations(t)
	})
}

// [VN] TestAuthService_RefreshToken kiểm thử việc cấp mới token qua refresh token
func TestAuthService_RefreshToken(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	publicID := "usr_token123"
	email := "refreshtoken@example.com"
	tokenVersion := 1

	validTokenPair, err := common.GenerateTokenPair(testJWTSecret, userID, publicID, email, tokenVersion)
	require.NoError(t, err, "failed to generate initial token pair for test")

	t.Run("Success", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewAuthService(mockUserRepo, mockCredRepo, mockProfileRepo, testJWTSecret)

		user := &models.User{
			ID:       userID,
			PublicID: publicID,
			Email:    email,
			Status:   models.UserStatusActive,
			Credentials: &models.UserCredentials{
				UserID:       userID,
				TokenVersion: tokenVersion,
			},
		}
		mockUserRepo.On("FindByID", ctx, userID).Return(user, nil)

		resp, err := service.RefreshToken(ctx, &dto.RefreshTokenRequest{
			RefreshToken: validTokenPair.RefreshToken,
		})

		require.NoError(t, err, "refresh token should succeed")
		require.NotNil(t, resp, "response should not be nil")
		assert.NotEmpty(t, resp.AccessToken, "new access token should be generated")
		assert.NotEmpty(t, resp.RefreshToken, "new refresh token should be generated")

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("InvalidOrExpiredToken", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewAuthService(mockUserRepo, mockCredRepo, mockProfileRepo, testJWTSecret)

		resp, err := service.RefreshToken(ctx, &dto.RefreshTokenRequest{
			RefreshToken: "invalid.jwt.token",
		})

		require.Error(t, err, "should return unauthorized for invalid token")
		assert.Nil(t, resp, "response should be nil")

		var appErr *common.DefaultError
		ok := errors.As(err, &appErr)
		require.True(t, ok, "error should be common.DefaultError")
		assert.Equal(t, 401, appErr.StatusCode(), "status should be 401")
	})

	t.Run("UserInactive", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewAuthService(mockUserRepo, mockCredRepo, mockProfileRepo, testJWTSecret)

		user := &models.User{
			ID:       userID,
			PublicID: publicID,
			Email:    email,
			Status:   models.UserStatusInactive,
		}
		mockUserRepo.On("FindByID", ctx, userID).Return(user, nil)

		resp, err := service.RefreshToken(ctx, &dto.RefreshTokenRequest{
			RefreshToken: validTokenPair.RefreshToken,
		})

		require.Error(t, err, "should return forbidden for inactive user")
		assert.Nil(t, resp, "response should be nil")

		var appErr *common.DefaultError
		ok := errors.As(err, &appErr)
		require.True(t, ok, "error should be common.DefaultError")
		assert.Equal(t, 403, appErr.StatusCode(), "status should be 403")

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("TokenRevoked_VersionMismatch", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewAuthService(mockUserRepo, mockCredRepo, mockProfileRepo, testJWTSecret)

		user := &models.User{
			ID:       userID,
			PublicID: publicID,
			Email:    email,
			Status:   models.UserStatusActive,
			Credentials: &models.UserCredentials{
				UserID:       userID,
				TokenVersion: 2,
			},
		}
		mockUserRepo.On("FindByID", ctx, userID).Return(user, nil)

		resp, err := service.RefreshToken(ctx, &dto.RefreshTokenRequest{
			RefreshToken: validTokenPair.RefreshToken,
		})

		require.Error(t, err, "should return unauthorized when session was revoked")
		assert.Nil(t, resp, "response should be nil")

		var appErr *common.DefaultError
		ok := errors.As(err, &appErr)
		require.True(t, ok, "error should be common.DefaultError")
		assert.Equal(t, 401, appErr.StatusCode(), "status should be 401")

		mockUserRepo.AssertExpectations(t)
	})
}

// [VN] TestAuthService_LogoutAll kiểm thử việc thu hồi phiên làm việc qua tăng token_version
func TestAuthService_LogoutAll(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewAuthService(mockUserRepo, mockCredRepo, mockProfileRepo, testJWTSecret)

		mockCredRepo.On("IncrementTokenVersion", ctx, (*gorm.DB)(nil), userID).Return(nil)

		err := service.LogoutAll(ctx, userID)
		require.NoError(t, err, "logout all should succeed")

		mockCredRepo.AssertExpectations(t)
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewAuthService(mockUserRepo, mockCredRepo, mockProfileRepo, testJWTSecret)

		mockCredRepo.On("IncrementTokenVersion", ctx, (*gorm.DB)(nil), userID).Return(errors.New("db error"))

		err := service.LogoutAll(ctx, userID)
		require.Error(t, err, "logout all should return error when db fails")

		var appErr *common.DefaultError
		ok := errors.As(err, &appErr)
		require.True(t, ok, "error should be common.DefaultError")
		assert.Equal(t, 500, appErr.StatusCode(), "status should be 500")

		mockCredRepo.AssertExpectations(t)
	})
}

// [VN] TestAuthService_GetMe kiểm thử tra cứu thông tin cá nhân của người dùng hiện hành
func TestAuthService_GetMe(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewAuthService(mockUserRepo, mockCredRepo, mockProfileRepo, testJWTSecret)

		user := &models.User{
			ID:       userID,
			PublicID: "usr_me123",
			Email:    "me@example.com",
			FullName: "My Name",
			Status:   models.UserStatusActive,
		}
		mockUserRepo.On("FindByID", ctx, userID).Return(user, nil)

		resp, err := service.GetMe(ctx, userID)
		require.NoError(t, err, "get me should succeed")
		require.NotNil(t, resp, "response should not be nil")
		assert.Equal(t, user.Email, resp.Email, "email should match")
		assert.Equal(t, user.FullName, resp.FullName, "fullname should match")

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCredRepo := new(MockCredentialRepository)
		mockProfileRepo := new(MockProfileRepository)

		service := services.NewAuthService(mockUserRepo, mockCredRepo, mockProfileRepo, testJWTSecret)

		mockUserRepo.On("FindByID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)

		resp, err := service.GetMe(ctx, userID)
		require.Error(t, err, "should return not found")
		assert.Nil(t, resp, "response should be nil")

		var appErr *common.DefaultError
		ok := errors.As(err, &appErr)
		require.True(t, ok, "error should be common.DefaultError")
		assert.Equal(t, 404, appErr.StatusCode(), "status should be 404")

		mockUserRepo.AssertExpectations(t)
	})
}
