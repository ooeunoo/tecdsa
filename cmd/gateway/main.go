package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"tecdsa/cmd/gateway/config"
	"tecdsa/cmd/gateway/server"
	"tecdsa/pkg/database"
	"tecdsa/pkg/database/repository"

	"gorm.io/gorm"
)

func main() {
	// 설정 로드
	cfg := loadConfig()

	// 전역 설정에 Bob과 Alice 주소 설정
	config.SetAddresses(cfg.BobGRPCAddress, cfg.AliceGRPCAddress)

	// 데이터베이스 연결
	db := connectDatabase(cfg)

	// 리포지토리 생성
	ipPublicKeyRepo := repository.NewClientSecurityRepository(db)

	// HTTP 서버 시작
	startHTTPServer(cfg, ipPublicKeyRepo)
}

func loadConfig() *config.Config {
	cfg := &config.Config{
		DBHost:           os.Getenv("DB_HOST"),
		DBUser:           os.Getenv("DB_USER"),
		DBPassword:       os.Getenv("DB_PASSWORD"),
		DBName:           os.Getenv("DB_NAME"),
		ServerPort:       os.Getenv("SERVER_PORT"),
		BobGRPCAddress:   os.Getenv("BOB_GRPC_ADDRESS"),
		AliceGRPCAddress: os.Getenv("ALICE_GRPC_ADDRESS"),
	}
	return cfg
}

func connectDatabase(cfg *config.Config) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBName)
	db, err := database.NewDatabase(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	// defer database.CloseDB(db)

	return db
}

func startHTTPServer(cfg *config.Config, ipPublicKeyRepo repository.ClientSecurityRepository) {
	srv := server.NewServer(cfg, ipPublicKeyRepo)

	log.Printf("Server listening on port %s", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, srv); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
