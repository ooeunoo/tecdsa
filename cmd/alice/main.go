package main

import (
	"log"

	"tecdsa/cmd/alice/server"
)

func main() {
	srv := server.NewServer()
	log.Fatal(srv.Run())
}
