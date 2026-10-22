package dto

import (
	"time"

	"github.com/vlahanam/terra-clone/models"
)

// [VN] CreateUserRequest định nghĩa dữ liệu đầu vào khi tạo mới người dùng
type CreateUserRequest struct {
	Email       string                 `json:"email" binding:"required,email"`
	Password    string                 `json:"password" binding:"required,min=8"`
	FullName    string                 `json:"full_name" binding:"required"`
	AvatarURL   *string                `json:"avatar_url,omitempty"`
	Status      string                 `json:"status,omitempty"`
	PhoneNumber *string                `json:"phone_number,omitempty"`
	DateOfBirth *string                `json:"date_of_birth,omitempty"`
	Gender      *string                `json:"gender,omitempty"`
	Address     *string                `json:"address,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
}

// [VN] UpdateUserRequest định nghĩa dữ liệu cập nhật thông tin người dùng
type UpdateUserRequest struct {
	FullName    *string                `json:"full_name,omitempty"`
	AvatarURL   *string                `json:"avatar_url,omitempty"`
	Status      *string                `json:"status,omitempty"`
	DateOfBirth *string                `json:"date_of_birth,omitempty"`
	Gender      *string                `json:"gender,omitempty"`
	PhoneNumber *string                `json:"phone_number,omitempty"`
	Address     *string                `json:"address,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
}

// [VN] ListUsersQuery định nghĩa tham số truy vấn tìm kiếm và phân trang cursor cho danh sách người dùng
type ListUsersQuery struct {
	Cursor string `form:"cursor"`
	Limit  int    `form:"limit"`
	Search string `form:"search"`
	Status string `form:"status"`
}

// [VN] UserProfileResponse định nghĩa dữ liệu hồ sơ cá nhân trả về cho client
type UserProfileResponse struct {
	DateOfBirth *string                `json:"date_of_birth,omitempty"`
	Gender      *string                `json:"gender,omitempty"`
	PhoneNumber *string                `json:"phone_number,omitempty"`
	Address     *string                `json:"address,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
}

// [VN] UserResponse định nghĩa thông tin người dùng hoàn chỉnh trả về cho client
type UserResponse struct {
	PublicID  string               `json:"id"`
	Email     string               `json:"email"`
	FullName  string               `json:"full_name"`
	AvatarURL *string              `json:"avatar_url,omitempty"`
	Status    string               `json:"status"`
	Profile   *UserProfileResponse `json:"profile,omitempty"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
}

// [VN] ToUserResponse chuyển đổi entity models.User sang DTO UserResponse
func ToUserResponse(user *models.User) *UserResponse {
	if user == nil {
		return nil
	}

	res := &UserResponse{
		PublicID:  user.PublicID,
		Email:     user.Email,
		FullName:  user.FullName,
		AvatarURL: user.AvatarURL,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	if user.Profile != nil {
		res.Profile = &UserProfileResponse{
			DateOfBirth: user.Profile.DateOfBirth,
			Gender:      user.Profile.Gender,
			PhoneNumber: user.Profile.PhoneNumber,
			Address:     user.Profile.Address,
			Preferences: user.Profile.Preferences,
		}
	}

	return res
}
