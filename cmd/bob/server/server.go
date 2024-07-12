package server

import (
	"log"
	handlers "tecdsa/cmd/bob/handlers"
	"tecdsa/pkg/database/repository"
	"tecdsa/pkg/service"

	pbKeygen "tecdsa/proto/keygen"
	pbSign "tecdsa/proto/sign"
)

type Server struct {
	pbKeygen.UnimplementedKeygenServiceServer
	pbSign.UnimplementedSignServiceServer
	keygenHandler  *handlers.KeygenHandler
	signHandler    *handlers.SignHandler
	networkService *service.NetworkService
}

func NewServer(repo repository.ParitalSecretShareRepository, networkService *service.NetworkService) *Server {
	return &Server{
		keygenHandler:  handlers.NewKeygenHandler(repo, networkService),
		signHandler:    handlers.NewSignHandler(repo),
		networkService: networkService,
	}
}

func (s *Server) KeyGen(stream pbKeygen.KeygenService_KeyGenServer) error {
	err := s.keygenHandler.HandleKeyGen(stream)
	if err != nil {
		log.Printf("Error in KeyGen: %v", err)
	}
	return err
}

func (s *Server) Sign(stream pbSign.SignService_SignServer) error {
	err := s.signHandler.HandleSign(stream)
	if err != nil {
		log.Printf("Error in Sign: %v", err)
	}
	return err
}
