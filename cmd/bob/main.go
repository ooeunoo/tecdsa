package main

import (
	"log"

	"tecdsa/cmd/bob/server"
)

func main() {
	srv := server.NewServer()
	log.Fatal(srv.Run())
}
