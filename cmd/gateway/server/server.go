package server

import (
	"net/http"

	"tecdsa/cmd/gateway/handlers"
)

func NewServer() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/key_gen", handlers.KeyGenHandler)
	return mux
}
