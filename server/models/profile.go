package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// JSONMap hỗ trợ lưu trữ và map trường preferences JSONB trong PostgreSQL
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	bytes, err := json.Marshal(j)
	if err != nil {
		return nil, fmt.Errorf("lỗi serialize JSONMap: %w", err)
	}
	return string(bytes), nil
}

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
		return errors.New("kiểu dữ liệu không tương thích với JSONMap")
	}

	result := make(JSONMap)
	if err := json.Unmarshal(bytes, &result); err != nil {
		return fmt.Errorf("lỗi deserialize JSONMap: %w", err)
	}
	*j = result
	return nil
}

type UserProfile struct {
	UserID      uuid.UUID  `gorm:"column:user_id;type:uuid;primaryKey" json:"user_id"`
	DateOfBirth *string    `gorm:"column:date_of_birth;type:date" json:"date_of_birth,omitempty"`
	Gender      *string    `gorm:"column:gender;type:varchar(20)" json:"gender,omitempty"`
	PhoneNumber *string    `gorm:"column:phone_number;type:varchar(20)" json:"phone_number,omitempty"`
	Address     *string    `gorm:"column:address;type:text" json:"address,omitempty"`
	Preferences JSONMap    `gorm:"column:preferences;type:jsonb" json:"preferences"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	User *User `gorm:"foreignKey:UserID;references:ID" json:"-"`
}

func (UserProfile) TableName() string {
	return "user_profiles"
}
