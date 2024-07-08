package database

import (
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var DB *gorm.DB

func InitDB() error {
	var err error
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dbConnectionString := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbName)

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
