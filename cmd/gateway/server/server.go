package server

import (
	"net/http"
	"path/filepath"

	"tecdsa/cmd/gateway/handlers"
)

func NewServer() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/key_gen", methodHandler(http.MethodPost, handlers.KeyGenHandler))
	mux.HandleFunc("/sign", methodHandler(http.MethodPost, handlers.SignHandler))
	mux.HandleFunc("/networks", methodHandler(http.MethodGet, handlers.GetAllNetworksHandler))
	mux.HandleFunc("/docs/", serveDoc)

	return mux

}

func methodHandler(method string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h(w, r)
	}
}

func serveDoc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	docPath := filepath.Join("cmd", "gateway", "docs", "index.html")
	http.ServeFile(w, r, docPath)
}
