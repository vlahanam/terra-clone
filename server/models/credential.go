package models

import (
	"time"

	"github.com/google/uuid"
)

type UserCredentials struct {
	UserID              uuid.UUID  `gorm:"column:user_id;type:uuid;primaryKey" json:"user_id"`
	PasswordHash        string     `gorm:"column:password_hash;type:varchar(255);not null" json:"-"`
	TokenVersion        int        `gorm:"column:token_version;not null;default:1" json:"token_version"`
	LastLoginAt         *time.Time `gorm:"column:last_login_at" json:"last_login_at,omitempty"`
	FailedLoginAttempts int        `gorm:"column:failed_login_attempts;not null;default:0" json:"failed_login_attempts"`
	LockedUntil         *time.Time `gorm:"column:locked_until" json:"locked_until,omitempty"`
	CreatedAt           time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	User *User `gorm:"foreignKey:UserID;references:ID" json:"-"`
}

func (UserCredentials) TableName() string {
	return "user_credentials"
}

func (c *UserCredentials) IsLocked() bool {
	if c.LockedUntil == nil {
		return false
	}
	return c.LockedUntil.After(time.Now())
}
