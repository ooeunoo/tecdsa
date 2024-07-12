package models

import "gorm.io/gorm"

type ParitalSecretShare struct {
	gorm.Model
	ID               uint32 `gorm:"primaryKey"`
	Address          string `gorm:"type:varchar(200);unique;not null"`
	Share            []byte `gorm:"type:blob;not null"`
	ClientSecurityID uint   `gorm:"index"`
}
