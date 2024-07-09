package models

import "gorm.io/gorm"

type Secret struct {
	gorm.Model
	Address         string `gorm:"type:varchar(42);unique;not null"`
	EncryptedSecret []byte `gorm:"type:blob;not null"`
	SecretKey       []byte `gorm:"type:blob;not null"`
}