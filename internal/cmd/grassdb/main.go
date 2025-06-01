package main

import (
	"grassdb/internal/server"
)

func main() {
	server.StartGRPCServer(":50051")
}
