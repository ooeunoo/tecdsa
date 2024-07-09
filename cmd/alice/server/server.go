// server.go
package server

import (
	"fmt"
	handlers "tecdsa/cmd/alice/handlers"
	"tecdsa/pkg/database/repository"

	pbKeygen "tecdsa/proto/keygen"
	pbSign "tecdsa/proto/sign"
)

type Server struct {
	pbKeygen.UnimplementedKeygenServiceServer
	pbSign.UnimplementedSignServiceServer
	keygenHandler *handlers.KeygenHandler
	signHandler   *handlers.SignHandler
}

func NewServer(repo repository.SecretRepository) *Server {
	return &Server{
		keygenHandler: handlers.NewKeygenHandler(repo),
		signHandler:   handlers.NewSignHandler(repo),
	}
}

func (s *Server) KeyGen(stream pbKeygen.KeygenService_KeyGenServer) error {
	fmt.Println("KeyGen called")
	return s.keygenHandler.HandleKeyGen(stream)
}

func (s *Server) Sign(stream pbSign.SignService_SignServer) error {
	return s.signHandler.HandleSign(stream)
}
