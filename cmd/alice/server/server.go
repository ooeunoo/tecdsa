// server.go
package server

import (
	handlers "tecdsa/cmd/alice/handlers"
	"tecdsa/pkg/database/repository"

	pbDkg "tecdsa/proto/dkg"
	pbSign "tecdsa/proto/sign"
)

type Server struct {
	pbDkg.UnimplementedDkgServiceServer
	pbSign.UnimplementedSignServiceServer
	dkgHandler  *handlers.DkgHandler
	signHandler *handlers.SignHandler
}

func NewServer(repo repository.SecretRepository) *Server {
	return &Server{
		dkgHandler: handlers.NewDkgHandler(repo),
		// signHandler: handlers.NewSignHandler(repo),
	}
}

func (s *Server) KeyGen(stream pbDkg.DkgService_KeyGenServer) error {
	return s.dkgHandler.HandleKeyGen(stream)
}
func (s *Server) Sign(stream pbSign.SignService_SignServer) error {
	return s.signHandler.HandleSign(stream)
}
