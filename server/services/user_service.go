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

// [VN] UserService định nghĩa các phương thức xử lý nghiệp vụ quản trị người dùng
type UserService interface {
	CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error)
	GetUserByID(ctx context.Context, idOrPublicID string) (*dto.UserResponse, error)
	ListUsers(ctx context.Context, query *dto.ListUsersQuery) ([]dto.UserResponse, *common.Paging, error)
	UpdateUser(ctx context.Context, idOrPublicID string, req *dto.UpdateUserRequest) (*dto.UserResponse, error)
	DeleteUser(ctx context.Context, idOrPublicID string, hard bool) error
}

type userService struct {
	userRepo    repositories.UserRepository
	credRepo    repositories.CredentialRepository
	profileRepo repositories.ProfileRepository
}

// [VN] NewUserService khởi tạo UserService với các repository tương ứng
func NewUserService(
	userRepo repositories.UserRepository,
	credRepo repositories.CredentialRepository,
	profileRepo repositories.ProfileRepository,
) UserService {
	return &userService{
		userRepo:    userRepo,
		credRepo:    credRepo,
		profileRepo: profileRepo,
	}
}

// [VN] CreateUser tạo mới người dùng kèm thông tin xác thực và hồ sơ cá nhân trong transaction
func (s *userService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	// [VN] Kiểm tra trùng email
	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, common.ErrConflict.
			WithReason("Địa chỉ email đã được sử dụng trên hệ thống").
			WithDetail("email", req.Email)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.ErrInternalServerError.
			WithReason("Lỗi khi kiểm tra thông tin email").
			WithTrace(err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, common.ErrInternalServerError.
			WithReason("Lỗi mã hóa mật khẩu").
			WithTrace(err)
	}

	status := models.UserStatusActive
	if req.Status != "" {
		status = req.Status
	}

	newUser := &models.User{
		PublicID:  common.GeneratePublicID("usr"),
		Email:     req.Email,
		FullName:  req.FullName,
		AvatarURL: req.AvatarURL,
		Status:    status,
	}

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
			DateOfBirth: req.DateOfBirth,
			Gender:      req.Gender,
			PhoneNumber: req.PhoneNumber,
			Address:     req.Address,
			Preferences: models.JSONMap{},
		}
		if req.Preferences != nil {
			profile.Preferences = req.Preferences
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
			WithReason("Không thể tạo mới người dùng").
			WithTrace(err)
	}

	return dto.ToUserResponse(newUser), nil
}

// [VN] GetUserByID tìm kiếm và trả về thông tin chi tiết người dùng theo ID hoặc public_id
func (s *userService) GetUserByID(ctx context.Context, idOrPublicID string) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByIDOrPublicID(ctx, idOrPublicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound.
				WithReason("Không tìm thấy người dùng được yêu cầu").
				WithDetail("id", idOrPublicID)
		}
		return nil, common.ErrInternalServerError.
			WithReason("Lỗi khi truy xuất dữ liệu người dùng").
			WithTrace(err)
	}

	return dto.ToUserResponse(user), nil
}

// [VN] ListUsers lấy danh sách người dùng theo điều kiện tìm kiếm, lọc trạng thái và phân trang cursor
func (s *userService) ListUsers(ctx context.Context, query *dto.ListUsersQuery) ([]dto.UserResponse, *common.Paging, error) {
	users, nextCursor, total, err := s.userRepo.List(ctx, query.Cursor, query.Limit, query.Search, query.Status)
	if err != nil {
		return nil, nil, common.ErrInternalServerError.
			WithReason("Lỗi khi lấy danh sách người dùng").
			WithTrace(err)
	}

	responses := make([]dto.UserResponse, 0, len(users))
	for i := range users {
		responses = append(responses, *dto.ToUserResponse(&users[i]))
	}

	paging := &common.Paging{
		Total:      total,
		FakeCursor: query.Cursor,
		NextCursor: nextCursor,
	}

	return responses, paging, nil
}

// [VN] UpdateUser cập nhật thông tin tài khoản và upsert thông tin hồ sơ cá nhân
func (s *userService) UpdateUser(ctx context.Context, idOrPublicID string, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByIDOrPublicID(ctx, idOrPublicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound.
				WithReason("Người dùng cần cập nhật không tồn tại").
				WithDetail("id", idOrPublicID)
		}
		return nil, common.ErrInternalServerError.
			WithReason("Lỗi khi tìm kiếm người dùng để cập nhật").
			WithTrace(err)
	}

	// [VN] Cập nhật trường thông tin cơ bản
	if req.FullName != nil {
		user.FullName = *req.FullName
	}
	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	user.UpdatedAt = time.Now()

	db := s.userRepo.GetDB()
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.userRepo.Update(ctx, tx, user); err != nil {
			return err
		}

		// [VN] Cập nhật profile nếu có các trường hồ sơ
		hasProfileUpdate := req.DateOfBirth != nil || req.Gender != nil || req.PhoneNumber != nil ||
			req.Address != nil || req.Preferences != nil

		if hasProfileUpdate {
			profile := user.Profile
			if profile == nil {
				profile = &models.UserProfile{
					UserID:      user.ID,
					Preferences: models.JSONMap{},
				}
			}

			if req.DateOfBirth != nil {
				profile.DateOfBirth = req.DateOfBirth
			}
			if req.Gender != nil {
				profile.Gender = req.Gender
			}
			if req.PhoneNumber != nil {
				profile.PhoneNumber = req.PhoneNumber
			}
			if req.Address != nil {
				profile.Address = req.Address
			}
			if req.Preferences != nil {
				profile.Preferences = req.Preferences
			}
			profile.UpdatedAt = time.Now()

			if err := s.profileRepo.Upsert(ctx, tx, profile); err != nil {
				return err
			}
			user.Profile = profile
		}

		return nil
	})

	if err != nil {
		return nil, common.ErrInternalServerError.
			WithReason("Lỗi cập nhật thông tin người dùng").
			WithTrace(err)
	}

	return dto.ToUserResponse(user), nil
}

// [VN] DeleteUser xóa người dùng (soft-delete đánh dấu status=deleted hoặc xóa cứng)
func (s *userService) DeleteUser(ctx context.Context, idOrPublicID string, hard bool) error {
	user, err := s.userRepo.FindByIDOrPublicID(ctx, idOrPublicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrNotFound.
				WithReason("Người dùng cần xóa không tồn tại").
				WithDetail("id", idOrPublicID)
		}
		return common.ErrInternalServerError.
			WithReason("Lỗi khi tìm kiếm người dùng").
			WithTrace(err)
	}

	parsedID, err := uuid.Parse(user.ID.String())
	if err != nil {
		return common.ErrInternalServerError.WithReason("Lỗi định dạng ID người dùng").WithTrace(err)
	}

	if err := s.userRepo.Delete(ctx, nil, parsedID, !hard); err != nil {
		return common.ErrInternalServerError.
			WithReason("Lỗi xóa người dùng khỏi hệ thống").
			WithTrace(err)
	}

	return nil
}
