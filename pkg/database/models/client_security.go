package models

import (
	"gorm.io/gorm"
)

type ClientSecurity struct {
	gorm.Model
	ID        uint32 `gorm:"primaryKey"`
	PublicKey string `gorm:"type:text;not null"`
	IP        string `gorm:"type:varchar(45);uniqueIndex"`
}
