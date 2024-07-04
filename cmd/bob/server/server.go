package server

import (
	"fmt"
	"net"

	"tecdsa/cmd/bob/handlers"
	pb "tecdsa/pkg/api/grpc/dkg"

	"google.golang.org/grpc"
)

type Server struct {
	grpcServer *grpc.Server
}

func NewServer() *Server {
	s := &Server{
		grpcServer: grpc.NewServer(),
	}
	pb.RegisterDKGServiceServer(s.grpcServer, &handlers.DKGHandler{})
	return s
}

func (s *Server) Run() error {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	return s.grpcServer.Serve(lis)
}
