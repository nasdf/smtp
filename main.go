package main

import (
	"net"

	"github.com/yondero/smtp/server"
	"github.com/yondero/smtp/storage"
)

func main() {
	storage := storage.NewS3Storage("yondero")
	
	ln, err := net.Listen("tcp", ":smtp")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}

		go server.NewSession(conn, storage).Serve()
	}
}
