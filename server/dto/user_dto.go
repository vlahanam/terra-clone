package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/vlahanam/terra-clone/models"
)

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

type ListUsersQuery struct {
	Cursor string `form:"cursor"`
	Limit  int    `form:"limit"`
	Search string `form:"search"`
	Status string `form:"status"`
}

type UserProfileResponse struct {
	DateOfBirth *string                `json:"date_of_birth,omitempty"`
	Gender      *string                `json:"gender,omitempty"`
	PhoneNumber *string                `json:"phone_number,omitempty"`
	Address     *string                `json:"address,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
}

type UserResponse struct {
	ID        uuid.UUID            `json:"id"`
	PublicID  string               `json:"public_id"`
	Email     string               `json:"email"`
	FullName  string               `json:"full_name"`
	AvatarURL *string              `json:"avatar_url,omitempty"`
	Status    string               `json:"status"`
	Profile   *UserProfileResponse `json:"profile,omitempty"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
}

func ToUserResponse(user *models.User) *UserResponse {
	if user == nil {
		return nil
	}

	res := &UserResponse{
		ID:        user.ID,
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
