package main

import (
	"log"
	"net"

	"tecdsa/cmd/bob/database"
	"tecdsa/cmd/bob/server"
	pb "tecdsa/pkg/api/grpc/dkg"

	"google.golang.org/grpc"
)

func main() {
	// Initialize database connection
	if err := database.InitDB(); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterDkgServiceServer(s, server.NewServer())
	log.Println("Bob server listening at :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
