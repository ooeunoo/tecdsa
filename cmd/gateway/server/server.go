package server

import (
	"net/http"

	"tecdsa/cmd/gateway/handlers"

	"github.com/gorilla/mux"
)

type Server struct {
	router *mux.Router
}

func NewServer() *Server {
	s := &Server{
		router: mux.NewRouter(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.router.HandleFunc("/dkg", handlers.DKGHandler).Methods("POST")
	s.router.HandleFunc("/refresh", handlers.RefreshHandler).Methods("POST")
	s.router.HandleFunc("/sign", handlers.SignHandler).Methods("POST")
}

func (s *Server) Run() error {
	return http.ListenAndServe(":8080", s.router)
}
