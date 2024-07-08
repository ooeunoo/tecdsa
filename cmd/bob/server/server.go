// server.go
package server

import (
	handlers "tecdsa/cmd/bob/handlers"

	pbDkg "tecdsa/pkg/api/grpc/dkg"
	pbSign "tecdsa/pkg/api/grpc/sign"
)

type Server struct {
	pbDkg.UnimplementedDkgServiceServer
	pbSign.UnimplementedSignServiceServer
	dkgHandler  *handlers.DkgHandler
	signHandler *handlers.SignHandler
}

func NewServer() *Server {
	return &Server{
		dkgHandler:  handlers.NewDkgHandler(),
		signHandler: handlers.NewSignHandler(),
	}
}

func (s *Server) KeyGen(stream pbDkg.DkgService_KeyGenServer) error {
	return s.dkgHandler.HandleKeyGen(stream)
}

func (s *Server) Sign(stream pbSign.SignService_SignServer) error {
	return s.signHandler.HandleSign(stream)
}
