package main

import (
	"log"

	"tecdsa/cmd/gateway/server"
)

func main() {
	srv := server.NewServer()
	log.Fatal(srv.Run())
}
