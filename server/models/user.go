package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// [VN] Các trạng thái tài khoản người dùng
const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
	UserStatusDeleted  = "deleted"
)

// [VN] User là entity đại diện cho bảng users trong CSDL
type User struct {
	ID          uuid.UUID        `gorm:"column:id;primaryKey" json:"id"`
	PublicID    string           `gorm:"column:public_id" json:"public_id"`
	Email       string           `gorm:"column:email" json:"email"`
	FullName    string           `gorm:"column:full_name" json:"full_name"`
	AvatarURL   *string          `gorm:"column:avatar_url" json:"avatar_url,omitempty"`
	Status      string           `gorm:"column:status" json:"status"`
	CreatedAt   time.Time        `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time        `gorm:"column:updated_at" json:"updated_at"`
	Credentials *UserCredentials `gorm:"foreignKey:UserID;references:ID" json:"credentials,omitempty"`
	Profile     *UserProfile     `gorm:"foreignKey:UserID;references:ID" json:"profile,omitempty"`
}

// [VN] TableName chỉ định tên bảng CSDL tương ứng với model User
func (User) TableName() string {
	return "users"
}

// [VN] BeforeCreate tự động sinh khóa chính UUIDv7 cho User trước khi chèn vào CSDL nếu chưa có ID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		newID, err := uuid.NewV7()
		if err != nil {
			return err
		}
		u.ID = newID
	}
	return nil
}

