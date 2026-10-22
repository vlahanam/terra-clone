package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vlahanam/terra-clone/common"
	"github.com/vlahanam/terra-clone/dto"
	"github.com/vlahanam/terra-clone/models"
	"github.com/vlahanam/terra-clone/repositories"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// [VN] Các hằng số quản lý chống tấn công Brute-force
const (
	// [VN] Số lần đăng nhập sai tối đa liên tiếp cho phép
	MaxLoginAttempts = 5
	// [VN] Thời gian tạm khóa tài khoản sau khi vượt quá số lần nhập sai
	AccountLockTime = 15 * time.Minute
)

// [VN] AuthService định nghĩa các nghiệp vụ xác thực người dùng
type AuthService interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error)
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.AuthResponse, error)
	LogoutAll(ctx context.Context, userID uuid.UUID) error
	GetMe(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error)
}

type authService struct {
	userRepo    repositories.UserRepository
	credRepo    repositories.CredentialRepository
	profileRepo repositories.ProfileRepository
	jwtSecret   string
}

// [VN] NewAuthService khởi tạo AuthService với các dependency cần thiết
func NewAuthService(
	userRepo repositories.UserRepository,
	credRepo repositories.CredentialRepository,
	profileRepo repositories.ProfileRepository,
	jwtSecret string,
) AuthService {
	return &authService{
		userRepo:    userRepo,
		credRepo:    credRepo,
		profileRepo: profileRepo,
		jwtSecret:   jwtSecret,
	}
}

// [VN] Register đăng ký người dùng mới trong transaction qua 3 bảng: users, user_credentials, user_profiles
func (s *authService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	// [VN] 1. Kiểm tra email đã tồn tại hay chưa
	existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, common.ErrConflict.
			WithReason("This email address is already registered in the system").
			WithDetail("email", req.Email)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.ErrInternalServerError.
			WithReason("Failed to verify user email").
			WithTrace(err)
	}

	// [VN] 2. Hash mật khẩu bằng bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, common.ErrInternalServerError.
			WithReason("Failed to hash password").
			WithTrace(err)
	}

	publicID := common.GeneratePublicID("usr")
	newUser := &models.User{
		PublicID: req.FullName, // [VN] Sẽ được gán lại đúng publicID bên dưới
		Email:    req.Email,
		FullName: req.FullName,
		Status:   models.UserStatusActive,
	}
	newUser.PublicID = publicID

	// [VN] 3. Thực thi Database Transaction khởi tạo 3 bảng
	db := s.userRepo.GetDB()
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.userRepo.Create(ctx, tx, newUser); err != nil {
			return err
		}

		creds := &models.UserCredentials{
			UserID:       newUser.ID,
			PasswordHash: string(hash),
			TokenVersion: 1,
		}
		if err := s.credRepo.Create(ctx, tx, creds); err != nil {
			return err
		}

		profile := &models.UserProfile{
			UserID:      newUser.ID,
			Preferences: models.JSONMap{},
		}
		if req.PhoneNumber != "" {
			profile.PhoneNumber = &req.PhoneNumber
		}
		if err := s.profileRepo.Create(ctx, tx, profile); err != nil {
			return err
		}

		newUser.Credentials = creds
		newUser.Profile = profile
		return nil
	})

	if err != nil {
		return nil, common.ErrInternalServerError.
			WithReason("Unable to complete user registration").
			WithTrace(err)
	}

	// [VN] 4. Phát sinh cặp JWT token
	tokens, err := common.GenerateTokenPair(s.jwtSecret, newUser.ID, newUser.PublicID, newUser.Email, 1)
	if err != nil {
		return nil, common.ErrInternalServerError.
			WithReason("Failed to generate authentication tokens").
			WithTrace(err)
	}

	return &dto.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresIn:    tokens.ExpiresIn,
		User:         dto.ToUserResponse(newUser),
	}, nil
}

// [VN] Login xác thực đăng nhập người dùng, kiểm tra khóa brute-force và phát sinh token
func (s *authService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	// [VN] 1. Tìm người dùng theo Email
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrUnauthorized.
				WithReason("Invalid email or password")
		}
		return nil, common.ErrInternalServerError.
			WithReason("Failed to query user information").
			WithTrace(err)
	}

	// [VN] 2. Kiểm tra trạng thái tài khoản
	if user.Status == models.UserStatusDeleted {
		return nil, common.ErrUnauthorized.
			WithReason("This account is no longer active in the system")
	}
	if user.Status == models.UserStatusInactive {
		return nil, common.ErrForbidden.
			WithReason("Your account has been deactivated, please contact system administrator")
	}

	// [VN] 3. Tra cứu credentials nếu chưa có sẵn
	creds := user.Credentials
	if creds == nil {
		creds, err = s.credRepo.FindByUserID(ctx, user.ID)
		if err != nil {
			return nil, common.ErrUnauthorized.
				WithReason("User credentials not found").
				WithTrace(err)
		}
	}

	// [VN] 4. Kiểm tra tài khoản có bị khóa tạm thời do brute-force
	if creds.IsLocked() {
		return nil, common.ErrForbidden.
			WithReasonf("Account is temporarily locked until %s due to multiple failed login attempts", creds.LockedUntil.Format("15:04:05 02/01/2006"))
	}

	// [VN] 5. So khớp mật khẩu băm
	if err := bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash), []byte(req.Password)); err != nil {
		// [VN] Nhập sai mật khẩu -> ghi nhận số lần sai và tự khóa nếu vượt ngưỡng
		updatedCreds, recErr := s.credRepo.RecordFailedLogin(ctx, nil, user.ID, MaxLoginAttempts, AccountLockTime)
		if recErr != nil {
			return nil, common.ErrInternalServerError.
				WithReason("Failed to update login status").
				WithTrace(recErr)
		}

		if updatedCreds.IsLocked() {
			return nil, common.ErrForbidden.
				WithReasonf("Too many failed login attempts (exceeded %d). Account is temporarily locked for 15 minutes until %s", MaxLoginAttempts, updatedCreds.LockedUntil.Format("15:04:05 02/01/2006"))
		}

		remainingAttempts := MaxLoginAttempts - updatedCreds.FailedLoginAttempts
		return nil, common.ErrUnauthorized.
			WithReasonf("Invalid email or password. You have %d attempts remaining", remainingAttempts)
	}

	// [VN] 6. Đăng nhập thành công -> Reset số lần nhập sai
	if err := s.credRepo.ResetFailedLogin(ctx, nil, user.ID); err != nil {
		return nil, common.ErrInternalServerError.
			WithReason("Failed to reset login session status").
			WithTrace(err)
	}

	// [VN] 7. Tạo JWT Token Pair
	tokens, err := common.GenerateTokenPair(s.jwtSecret, user.ID, user.PublicID, user.Email, creds.TokenVersion)
	if err != nil {
		return nil, common.ErrInternalServerError.
			WithReason("Failed to generate authentication tokens").
			WithTrace(err)
	}

	return &dto.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresIn:    tokens.ExpiresIn,
		User:         dto.ToUserResponse(user),
	}, nil
}

// [VN] RefreshToken cấp lại Access Token mới dựa trên Refresh Token còn hạn
func (s *authService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.AuthResponse, error) {
	// [VN] 1. Kiểm tra tính hợp lệ và loại token
	claims, err := common.ValidateToken(s.jwtSecret, req.RefreshToken, common.TokenTypeRefresh)
	if err != nil {
		return nil, common.ErrUnauthorized.
			WithReason("Refresh token is invalid or expired").
			WithTrace(err)
	}

	// [VN] 2. Tra cứu người dùng và thông tin credential
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, common.ErrUnauthorized.
			WithReason("Token owner does not exist").
			WithTrace(err)
	}

	if user.Status != models.UserStatusActive {
		return nil, common.ErrForbidden.
			WithReason("User account is not currently active")
	}

	creds := user.Credentials
	if creds == nil {
		creds, err = s.credRepo.FindByUserID(ctx, user.ID)
		if err != nil {
			return nil, common.ErrUnauthorized.
				WithReason("User credentials not found").
				WithTrace(err)
		}
	}

	// [VN] 3. Kiểm tra token_version xem token này đã bị thu hồi chưa
	if creds.TokenVersion != claims.TokenVersion {
		return nil, common.ErrUnauthorized.
			WithReason("This session has been revoked due to logout or password change")
	}

	// [VN] 4. Phát sinh Token Pair mới
	tokens, err := common.GenerateTokenPair(s.jwtSecret, user.ID, user.PublicID, user.Email, creds.TokenVersion)
	if err != nil {
		return nil, common.ErrInternalServerError.
			WithReason("Failed to generate new token pair").
			WithTrace(err)
	}

	return &dto.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresIn:    tokens.ExpiresIn,
		User:         dto.ToUserResponse(user),
	}, nil
}

// [VN] LogoutAll tăng token_version để thu hồi toàn bộ các phiên đăng nhập của người dùng
func (s *authService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	if err := s.credRepo.IncrementTokenVersion(ctx, nil, userID); err != nil {
		return common.ErrInternalServerError.
			WithReason("Failed to revoke user sessions").
			WithTrace(err)
	}
	return nil
}

// [VN] GetMe lấy thông tin chi tiết người dùng đang đăng nhập
func (s *authService) GetMe(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound.
				WithReason("User information not found").
				WithTrace(err)
		}
		return nil, common.ErrInternalServerError.
			WithReason("Failed to query user information").
			WithTrace(err)
	}

	return dto.ToUserResponse(user), nil
}
