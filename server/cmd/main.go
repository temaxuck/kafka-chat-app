package main

import (
	"flag"

	"server/pkg/server"
)

var addr = flag.String("addr", "localhost:8080", "HTTP server address")

func main() {
	flag.Parse()

	server.Run(addr)
}
