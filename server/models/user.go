package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
	UserStatusDeleted  = "deleted"
)

type User struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	PublicID  string     `gorm:"column:public_id;type:varchar(50);unique;not null" json:"public_id"`
	Email     string     `gorm:"column:email;type:varchar(255);unique;not null" json:"email"`
	FullName  string     `gorm:"column:full_name;type:varchar(100);not null" json:"full_name"`
	AvatarURL *string    `gorm:"column:avatar_url;type:text" json:"avatar_url,omitempty"`
	Status    string     `gorm:"column:status;type:varchar(20);not null;default:'active'" json:"status"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	Credentials *UserCredentials `gorm:"foreignKey:UserID;references:ID" json:"credentials,omitempty"`
	Profile     *UserProfile     `gorm:"foreignKey:UserID;references:ID" json:"profile,omitempty"`
}

func (User) TableName() string {
	return "users"
}
