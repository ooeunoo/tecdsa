package main

import (
	"log"
	"net"

	"tecdsa/cmd/alice/server"
	pb "tecdsa/pkg/api/grpc/dkg"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterDkgServiceServer(s, server.NewServer())
	log.Println("Alice server listening at :50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
