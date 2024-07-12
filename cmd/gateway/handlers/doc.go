package handlers

import (
	"net/http"
	"path/filepath"
)

type ServeDocHandler struct {
	docRoot string
}

func NewServeDocHandler(docRoot string) *ServeDocHandler {
	return &ServeDocHandler{
		docRoot: docRoot,
	}
}

func (h *ServeDocHandler) Serve(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/docs" || r.URL.Path == "/docs/" {
		http.ServeFile(w, r, filepath.Join(h.docRoot, "index.html"))
		return
	}

	http.FileServer(http.Dir(h.docRoot)).ServeHTTP(w, r)
}
