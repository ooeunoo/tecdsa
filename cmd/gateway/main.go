package main

import (
	"log"
	"net/http"

	"tecdsa/cmd/gateway/server"
)

func main() {
	// db.Init()
	srv := server.NewServer()
	log.Fatal(http.ListenAndServe(":8080", srv))
}
