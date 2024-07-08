package database

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var DB *gorm.DB

func InitDB() error {
	var err error
	dbConnectionString := "user:password@tcp(127.0.0.1:3306)/gateway?charset=utf8mb4&parseTime=True&loc=Local"
	DB, err = gorm.Open("mysql", dbConnectionString)
	if err != nil {
		return fmt.Errorf("failed to connect database: %v", err)
	}

	// Auto Migrate
	DB.AutoMigrate(&Secret{})

	return nil
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}
