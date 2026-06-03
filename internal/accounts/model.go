package accounts

import (
	"time"

	"gorm.io/gorm"
)


const CtxUserKey = "currentUser"

type User struct{
	gorm.Model
	Email string `json:"email" gorm:"column:email;not null"`
	DisplayName string `json:"display_name" gorm:"column:display_name;not null"`
	IsActive bool `json:"is_active" gorm:"column:is_active;not null"`
	IsStaff bool `json:"is_staff" gorm:"column:is_staff;not null"`
  	DateJoined time.Time `gorm:"autoCreateTime"`

}