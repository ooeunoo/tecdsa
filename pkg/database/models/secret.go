package models

import "gorm.io/gorm"

type Secret struct {
	gorm.Model
	Address   string `gorm:"type:varchar(200);unique;not null"`
	Share     []byte `gorm:"type:blob;not null"`
	SecretKey []byte `gorm:"type:blob;not null"`
}
