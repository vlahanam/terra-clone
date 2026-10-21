package models

import (
	"time"

	"github.com/google/uuid"
)

// [VN] UserCredentials là entity quản lý thông tin đăng nhập, bảo mật và token của người dùng
type UserCredentials struct {
	UserID              uuid.UUID  `gorm:"column:user_id;primaryKey" json:"user_id"`
	PasswordHash        string     `gorm:"column:password_hash" json:"-"`
	TokenVersion        int        `gorm:"column:token_version" json:"token_version"`
	LastLoginAt         *time.Time `gorm:"column:last_login_at" json:"last_login_at,omitempty"`
	FailedLoginAttempts int        `gorm:"column:failed_login_attempts" json:"failed_login_attempts"`
	LockedUntil         *time.Time `gorm:"column:locked_until" json:"locked_until,omitempty"`
	CreatedAt           time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at" json:"updated_at"`
	User                *User      `gorm:"foreignKey:UserID;references:ID" json:"-"`
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
