// server/server.go
package server

import (
	"net/http"
	"path/filepath"

	"tecdsa/cmd/gateway/config"
	"tecdsa/cmd/gateway/handlers"

	createUnsignedTxHandlers "tecdsa/cmd/gateway/handlers/create_unsigned_tx"
	"tecdsa/pkg/database/repository"
	"tecdsa/pkg/service"
)

type Server struct {
	clientSecurityRepo repository.ClientSecurityRepository
	mux                *http.ServeMux
	config             *config.Config
	networkService     *service.NetworkService
}

func NewServer(cfg *config.Config, clientSecurityRepo repository.ClientSecurityRepository) *Server {
	s := &Server{
		clientSecurityRepo: clientSecurityRepo,
		mux:                http.NewServeMux(),
		config:             cfg,
		networkService:     service.NewNetworkService(),
	}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("/register", s.methodHandler(http.MethodPost, s.registerClientSecurityHandler()))
	s.mux.HandleFunc("/key_gen", s.methodHandler(http.MethodPost, s.keyGenHandler()))
	s.mux.HandleFunc("/sign", s.methodHandler(http.MethodPost, s.signHandler()))
	s.mux.HandleFunc("/networks", s.methodHandler(http.MethodGet, s.getAllNetworksHandler()))
	s.mux.HandleFunc("/create_unsigned_tx/", s.methodHandler(http.MethodPost, s.createUnsignedTxHandler()))
	s.mux.HandleFunc("/docs/", s.methodHandler(http.MethodGet, s.serveDocHandler()))

}

func (s *Server) methodHandler(method string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h(w, r)
	}
}

func (s *Server) registerClientSecurityHandler() http.HandlerFunc {
	handler := handlers.NewRegisterClientSecurityHandler(s.clientSecurityRepo)
	return handler.Serve
}

func (s *Server) keyGenHandler() http.HandlerFunc {
	handler := handlers.NewKeyGenHandler(s.config, s.clientSecurityRepo, s.networkService)
	return handler.Serve
}

func (s *Server) signHandler() http.HandlerFunc {
	handler := handlers.NewSignHandler(s.config, s.clientSecurityRepo, s.networkService)
	return handler.Serve
}

func (s *Server) getAllNetworksHandler() http.HandlerFunc {
	handler := handlers.NewGetAllNetworksHandler(s.networkService)
	return handler.Serve
}

func (s *Server) serveDocHandler() http.HandlerFunc {
	docRoot := filepath.Join("cmd", "gateway", "docs") // Adjust this path as needed
	handler := handlers.NewServeDocHandler(docRoot)
	return handler.Serve
}
func (s *Server) createUnsignedTxHandler() http.HandlerFunc {
	handler := createUnsignedTxHandlers.NewCreateUnsignedTxHandler(s.networkService)
	return handler.Serve
}
