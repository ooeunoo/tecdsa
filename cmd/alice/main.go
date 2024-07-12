package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"tecdsa/cmd/alice/config"
	"tecdsa/cmd/alice/server"
	"tecdsa/pkg/database"
	"tecdsa/pkg/database/repository"
	"tecdsa/pkg/service"

	pbKeygen "tecdsa/proto/keygen"
	pbSign "tecdsa/proto/sign"

	"google.golang.org/grpc"
	"gorm.io/gorm"
)

func main() {
	// 설정 로드
	cfg := loadConfig()

	// 데이터베이스 연결
	db := connectDatabase(cfg)

	// 리포지토리 생성
	paritalSecretShareRepository := repository.NewPartialSecretShareRepository(db)

	// 네트워크 서비스 생성
	networkService := service.NewNetworkService()

	// gRPC 서버 시작
	startGRPCServer(cfg, paritalSecretShareRepository, networkService)
}

func loadConfig() *config.Config {
	cfg := &config.Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		ServerPort: os.Getenv("SERVER_PORT"),
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

func startGRPCServer(cfg *config.Config, repo repository.ParitalSecretShareRepository, networkService *service.NetworkService) {
	lis, err := net.Listen("tcp", ":"+cfg.ServerPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	srv := server.NewServer(repo, networkService)

	pbKeygen.RegisterKeygenServiceServer(s, srv)
	pbSign.RegisterSignServiceServer(s, srv)

	log.Printf("Alice server listening at :%s", cfg.ServerPort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
