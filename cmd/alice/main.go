package main

import (
	"context"
	"log"
	"net"

	"tecdsa/internal/encoding"
	pb "tecdsa/pkg/api/grpc/dkg"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedDKGServiceServer
}

func (s *server) ProcessDKG(ctx context.Context, req *pb.DKGRequest) (*pb.DKGResponse, error) {
	decodedNumber := encoding.Decode(req.RandomNumber)
	result := decodedNumber * 2
	encodedResult := encoding.Encode(result)

	return &pb.DKGResponse{Result: encodedResult}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterDKGServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
