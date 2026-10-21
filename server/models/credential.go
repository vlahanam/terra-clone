package models

import (
	"time"

	"github.com/google/uuid"
)

// [VN] UserCredentials là entity quản lý thông tin đăng nhập, bảo mật và token của người dùng
type UserCredentials struct {
	UserID uuid.UUID `gorm:"column:user_id;type:uuid;primaryKey" json:"user_id"`
	PasswordHash string `gorm:"column:password_hash;type:varchar(255);not null" json:"-"`
	TokenVersion int `gorm:"column:token_version;not null;default:1" json:"token_version"`
	LastLoginAt *time.Time `gorm:"column:last_login_at" json:"last_login_at,omitempty"`
	FailedLoginAttempts int `gorm:"column:failed_login_attempts;not null;default:0" json:"failed_login_attempts"`
	LockedUntil *time.Time `gorm:"column:locked_until" json:"locked_until,omitempty"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	User *User `gorm:"foreignKey:UserID;references:ID" json:"-"`
}

// [VN] TableName chỉ định tên bảng CSDL tương ứng với model UserCredentials
func (UserCredentials) TableName() string {
	return "user_credentials"
}

// [VN] IsLocked kiểm tra tài khoản có đang trong thời gian bị khóa tạm thời hay không
func (c *UserCredentials) IsLocked() bool {
	if c.LockedUntil == nil {
		return false
	}
	return c.LockedUntil.After(time.Now())
}
