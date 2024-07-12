package database

import (
	"fmt"

	"tecdsa/pkg/database/models"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDatabase(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %v", err)
	}

	// Auto Migrate
	db.AutoMigrate(&models.ParitalSecretShare{}, &models.ClientSecurity{})

	return db, nil
}

func CloseDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		fmt.Printf("Failed to get sql.DB from gorm.DB: %v\n", err)
	}
	return sqlDB.Close()
}
