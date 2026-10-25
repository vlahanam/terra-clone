package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/vlahanam/terra-clone/common"
	"gorm.io/gorm"
)

// [VN] Các vai trò thành viên trong công ty
const (
	UserCompanyRoleOwner  = "owner"
	UserCompanyRoleAdmin  = "admin"
	UserCompanyRoleMember = "member"
	UserCompanyRoleGuest  = "guest"
)

// [VN] Các trạng thái liên kết giữa người dùng và công ty
const (
	UserCompanyStatusActive    = "active"
	UserCompanyStatusInactive  = "inactive"
	UserCompanyStatusPending   = "pending"
	UserCompanyStatusSuspended = "suspended"
	UserCompanyStatusDeleted   = "deleted"
)

// [VN] UserCompany là entity đại diện cho bảng user_companies trong CSDL
type UserCompany struct {
	ID          uuid.UUID  `gorm:"column:id;primaryKey" json:"id"`
	PublicID    string     `gorm:"column:public_id" json:"public_id"`
	UserID      uuid.UUID  `gorm:"column:user_id" json:"user_id"`
	CompanyID   uuid.UUID  `gorm:"column:company_id" json:"company_id"`
	Role        string     `gorm:"column:role" json:"role"`
	JobTitle    *string    `gorm:"column:job_title" json:"job_title,omitempty"`
	Department  *string    `gorm:"column:department" json:"department,omitempty"`
	Status      string     `gorm:"column:status" json:"status"`
	Permissions JSONMap    `gorm:"column:permissions" json:"permissions"`
	Metadata    JSONMap    `gorm:"column:metadata" json:"metadata"`
	JoinedAt    *time.Time `gorm:"column:joined_at" json:"joined_at,omitempty"`
	CreatedAt   time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at" json:"updated_at"`
	User        *User      `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Company     *Company   `gorm:"foreignKey:CompanyID;references:ID" json:"company,omitempty"`
}

// [VN] TableName chỉ định tên bảng CSDL tương ứng với model UserCompany
func (UserCompany) TableName() string {
	return "user_companies"
}

// [VN] BeforeCreate tự động sinh khóa chính UUIDv7 và public_id cho UserCompany trước khi chèn vào CSDL
func (uc *UserCompany) BeforeCreate(tx *gorm.DB) error {
	if uc.ID == uuid.Nil {
		newID, err := uuid.NewV7()
		if err != nil {
			return err
		}
		uc.ID = newID
	}
	if uc.PublicID == "" {
		uc.PublicID = common.GeneratePublicID("ucmp")
	}
	if uc.Role == "" {
		uc.Role = UserCompanyRoleMember
	}
	if uc.Status == "" {
		uc.Status = UserCompanyStatusActive
	}
	if uc.Permissions == nil {
		uc.Permissions = make(JSONMap)
	}
	if uc.Metadata == nil {
		uc.Metadata = make(JSONMap)
	}
	return nil
}
