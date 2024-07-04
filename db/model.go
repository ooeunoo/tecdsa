package db

import (
	"gorm.io/gorm"
)

type RequestLog struct {
	gorm.Model
	Path      string
	Region    string
	Params    string `gorm:"type:text"`
	Body      string `gorm:"type:text"`
	Method    string
	Headers   string `gorm:"type:text"`
	IPAddress string
	Response  string `gorm:"type:text"` // 응답 결과를 저장할 필드 추가
}
