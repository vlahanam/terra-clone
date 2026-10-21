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

const (
	MaxLoginAttempts = 5
	AccountLockTime  = 15 * time.Minute
)

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

func (s *authService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	// 1. Kiểm tra email đã tồn tại hay chưa
	existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, common.ErrConflict.
			WithReason("Địa chỉ email này đã được đăng ký trong hệ thống").
			WithDetail("email", req.Email)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.ErrInternalServerError.
			WithReason("Lỗi khi kiểm tra email người dùng").
			WithTrace(err)
	}

	// 2. Hash mật khẩu bằng bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, common.ErrInternalServerError.
			WithReason("Lỗi mã hóa mật khẩu").
			WithTrace(err)
	}

	publicID := common.GeneratePublicID("usr")
	newUser := &models.User{
		PublicID: req.FullName, // sẽ được gán lại đúng publicID bên dưới
		Email:    req.Email,
		FullName: req.FullName,
		Status:   models.UserStatusActive,
	}
	newUser.PublicID = publicID

	// 3. Thực thi Database Transaction khởi tạo 3 bảng
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
			WithReason("Không thể hoàn tất đăng ký tài khoản").
			WithTrace(err)
	}

	// 4. Phát sinh cặp JWT token
	tokens, err := common.GenerateTokenPair(s.jwtSecret, newUser.ID, newUser.PublicID, newUser.Email, 1)
	if err != nil {
		return nil, common.ErrInternalServerError.
			WithReason("Lỗi phát sinh token xác thực").
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

func (s *authService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	// 1. Tìm người dùng theo Email
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrUnauthorized.
				WithReason("Email hoặc mật khẩu không chính xác")
		}
		return nil, common.ErrInternalServerError.
			WithReason("Lỗi tra cứu thông tin người dùng").
			WithTrace(err)
	}

	// 2. Kiểm tra trạng thái tài khoản
	if user.Status == models.UserStatusDeleted {
		return nil, common.ErrUnauthorized.
			WithReason("Tài khoản này không còn hoạt động trên hệ thống")
	}
	if user.Status == models.UserStatusInactive {
		return nil, common.ErrForbidden.
			WithReason("Tài khoản của bạn đã bị vô hiệu hóa, vui lòng liên hệ quản trị viên")
	}

	// 3. Tra cứu credentials nếu chưa có sẵn
	creds := user.Credentials
	if creds == nil {
		creds, err = s.credRepo.FindByUserID(ctx, user.ID)
		if err != nil {
			return nil, common.ErrUnauthorized.
				WithReason("Thông tin xác thực không tồn tại").
				WithTrace(err)
		}
	}

	// 4. Kiểm tra tài khoản có bị khóa tạm thời do brute-force
	if creds.IsLocked() {
		return nil, common.ErrForbidden.
			WithReasonf("Tài khoản đang bị tạm khóa đến %s do nhập sai mật khẩu nhiều lần", creds.LockedUntil.Format("15:04:05 02/01/2006"))
	}

	// 5. So khớp mật khẩu băm
	if err := bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash), []byte(req.Password)); err != nil {
		// Nhập sai mật khẩu -> ghi nhận số lần sai và tự khóa nếu vượt ngưỡng
		updatedCreds, recErr := s.credRepo.RecordFailedLogin(ctx, nil, user.ID, MaxLoginAttempts, AccountLockTime)
		if recErr != nil {
			return nil, common.ErrInternalServerError.
				WithReason("Lỗi cập nhật trạng thái đăng nhập").
				WithTrace(recErr)
		}

		if updatedCreds.IsLocked() {
			return nil, common.ErrForbidden.
				WithReasonf("Bạn đã nhập sai quá %d lần. Tài khoản tạm khóa 15 phút đến %s", MaxLoginAttempts, updatedCreds.LockedUntil.Format("15:04:05 02/01/2006"))
		}

		remainingAttempts := MaxLoginAttempts - updatedCreds.FailedLoginAttempts
		return nil, common.ErrUnauthorized.
			WithReasonf("Email hoặc mật khẩu không chính xác. Bạn còn %d lần thử", remainingAttempts)
	}

	// 6. Đăng nhập thành công -> Reset số lần nhập sai
	if err := s.credRepo.ResetFailedLogin(ctx, nil, user.ID); err != nil {
		return nil, common.ErrInternalServerError.
			WithReason("Lỗi làm mới trạng thái phiên đăng nhập").
			WithTrace(err)
	}

	// 7. Tạo JWT Token Pair
	tokens, err := common.GenerateTokenPair(s.jwtSecret, user.ID, user.PublicID, user.Email, creds.TokenVersion)
	if err != nil {
		return nil, common.ErrInternalServerError.
			WithReason("Lỗi tạo token xác thực").
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

func (s *authService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.AuthResponse, error) {
	// 1. Kiểm tra tính hợp lệ và loại token
	claims, err := common.ValidateToken(s.jwtSecret, req.RefreshToken, common.TokenTypeRefresh)
	if err != nil {
		return nil, common.ErrUnauthorized.
			WithReason("Refresh token không hợp lệ hoặc đã hết hạn").
			WithTrace(err)
	}

	// 2. Tra cứu người dùng và thông tin credential
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, common.ErrUnauthorized.
			WithReason("Người dùng của token không tồn tại").
			WithTrace(err)
	}

	if user.Status != models.UserStatusActive {
		return nil, common.ErrForbidden.
			WithReason("Tài khoản người dùng hiện không ở trạng thái hoạt động")
	}

	creds := user.Credentials
	if creds == nil {
		creds, err = s.credRepo.FindByUserID(ctx, user.ID)
		if err != nil {
			return nil, common.ErrUnauthorized.
				WithReason("Không tìm thấy thông tin xác thực").
				WithTrace(err)
		}
	}

	// 3. Kiểm tra token_version xem token này đã bị thu hồi chưa
	if creds.TokenVersion != claims.TokenVersion {
		return nil, common.ErrUnauthorized.
			WithReason("Phiên đăng nhập này đã bị thu hồi do đăng xuất hoặc đổi mật khẩu")
	}

	// 4. Phát sinh Token Pair mới
	tokens, err := common.GenerateTokenPair(s.jwtSecret, user.ID, user.PublicID, user.Email, creds.TokenVersion)
	if err != nil {
		return nil, common.ErrInternalServerError.
			WithReason("Lỗi phát sinh token mới").
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

func (s *authService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	if err := s.credRepo.IncrementTokenVersion(ctx, nil, userID); err != nil {
		return common.ErrInternalServerError.
			WithReason("Lỗi thu hồi các phiên đăng nhập").
			WithTrace(err)
	}
	return nil
}

func (s *authService) GetMe(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound.
				WithReason("Không tìm thấy thông tin người dùng").
				WithTrace(err)
		}
		return nil, common.ErrInternalServerError.
			WithReason("Lỗi truy vấn thông tin người dùng").
			WithTrace(err)
	}

	return dto.ToUserResponse(user), nil
}
