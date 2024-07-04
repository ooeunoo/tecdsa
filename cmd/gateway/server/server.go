package server

import (
	"net/http"

	"tecdsa/cmd/gateway/handlers"
	"tecdsa/db"
)

type Server struct {
	router *http.ServeMux
}

func NewServer() *Server {
	s := &Server{
		router: http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.router.HandleFunc("/dkg", db.LogRequest(handlers.DKGHandler))
}

func (s *Server) Run() error {
	return http.ListenAndServe(":8080", s.router)
}
