package main

import (
	"log"

	"tecdsa/cmd/gateway/server"
	"tecdsa/db"
)

func main() {
	db.Init()
	srv := server.NewServer()
	log.Fatal(srv.Run())
}
