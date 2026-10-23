package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// [VN] JSONMap hỗ trợ lưu trữ và map trường preferences JSONB trong PostgreSQL
type JSONMap map[string]interface{}

// [VN] Value chuyển đổi JSONMap sang driver.Value để lưu trữ vào trường JSONB PostgreSQL
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	bytes, err := json.Marshal(j)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize JSONMap: %w", err)
	}
	return string(bytes), nil
}

// [VN] Scan đọc dữ liệu từ trường JSONB PostgreSQL và deserialize vào JSONMap
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONMap)
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("incompatible type for JSONMap")
	}

	result := make(JSONMap)
	if err := json.Unmarshal(bytes, &result); err != nil {
		return fmt.Errorf("failed to deserialize JSONMap: %w", err)
	}
	*j = result
	return nil
}

// [VN] UserProfile là entity đại diện cho bảng user_profiles lưu thông tin mở rộng của người dùng
type UserProfile struct {
	UserID          uuid.UUID `gorm:"column:user_id;primaryKey" json:"user_id"`
	DateOfBirth     *string   `gorm:"column:date_of_birth" json:"date_of_birth,omitempty"`
	Gender          *string   `gorm:"column:gender" json:"gender,omitempty"`
	PhoneNumber     *string   `gorm:"column:phone_number" json:"phone_number,omitempty"`
	Address         *string   `gorm:"column:address" json:"address,omitempty"`
	CountryCode     *string   `gorm:"column:country_code" json:"country_code,omitempty"`
	Nationality     *string   `gorm:"column:nationality" json:"nationality,omitempty"`
	Timezone        *string   `gorm:"column:timezone" json:"timezone,omitempty"`
	Locale          *string   `gorm:"column:locale" json:"locale,omitempty"`
	City            *string   `gorm:"column:city" json:"city,omitempty"`
	StateOrProvince *string   `gorm:"column:state_or_province" json:"state_or_province,omitempty"`
	PostalCode      *string   `gorm:"column:postal_code" json:"postal_code,omitempty"`
	Preferences     JSONMap   `gorm:"column:preferences" json:"preferences"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updated_at"`
	User            *User     `gorm:"foreignKey:UserID;references:ID" json:"-"`
}

// [VN] TableName chỉ định tên bảng CSDL tương ứng với model UserProfile
func (UserProfile) TableName() string {
	return "user_profiles"
}
