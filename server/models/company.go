package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// [VN] Các trạng thái hoạt động của công ty
const (
	CompanyStatusActive    = "active"
	CompanyStatusInactive  = "inactive"
	CompanyStatusSuspended = "suspended"
	CompanyStatusDeleted   = "deleted"
)

// [VN] Company là entity đại diện cho bảng companies trong CSDL
type Company struct {
	ID              uuid.UUID  `gorm:"column:id;primaryKey" json:"id"`
	PublicID        string     `gorm:"column:public_id" json:"public_id"`
	Code            string     `gorm:"column:code" json:"code"`
	Name            string     `gorm:"column:name" json:"name"`
	LegalName       *string    `gorm:"column:legal_name" json:"legal_name,omitempty"`
	TaxID           *string    `gorm:"column:tax_id" json:"tax_id,omitempty"`
	Email           *string    `gorm:"column:email" json:"email,omitempty"`
	Phone           *string    `gorm:"column:phone" json:"phone,omitempty"`
	Address         *string    `gorm:"column:address" json:"address,omitempty"`
	City            *string    `gorm:"column:city" json:"city,omitempty"`
	StateOrProvince *string    `gorm:"column:state_or_province" json:"state_or_province,omitempty"`
	PostalCode      *string    `gorm:"column:postal_code" json:"postal_code,omitempty"`
	CountryCode     *string    `gorm:"column:country_code" json:"country_code,omitempty"`
	Timezone        *string    `gorm:"column:timezone" json:"timezone,omitempty"`
	Currency        *string    `gorm:"column:currency" json:"currency,omitempty"`
	Website         *string    `gorm:"column:website" json:"website,omitempty"`
	LogoURL         *string    `gorm:"column:logo_url" json:"logo_url,omitempty"`
	Description     *string    `gorm:"column:description" json:"description,omitempty"`
	Status          string     `gorm:"column:status" json:"status"`
	OwnerID         *uuid.UUID `gorm:"column:owner_id" json:"owner_id,omitempty"`
	Metadata        JSONMap    `gorm:"column:metadata" json:"metadata"`
	CreatedAt       time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at" json:"updated_at"`
	Owner           *User      `gorm:"foreignKey:OwnerID;references:ID" json:"owner,omitempty"`
}

// [VN] TableName chỉ định tên bảng CSDL tương ứng với model Company
func (Company) TableName() string {
	return "companies"
}

// [VN] BeforeCreate tự động sinh khóa chính UUIDv7 cho Company trước khi chèn vào CSDL nếu chưa có ID
func (c *Company) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		newID, err := uuid.NewV7()
		if err != nil {
			return err
		}
		c.ID = newID
	}
	return nil
}
