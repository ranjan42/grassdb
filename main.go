// main.go
package main

import (
	"flag"
	"strings"

	"grassdb/internal/raft"
	"grassdb/internal/server"
)

func main() {
	id := flag.String("id", "node1", "Unique node ID")
	addr := flag.String("addr", ":50051", "Address to listen on for gRPC")
	httpAddr := flag.String("http", ":8080", "Address to listen on for HTTP")
	peersStr := flag.String("peers", "", "Comma-separated list of peer addresses (e.g. 127.0.0.1:50052,127.0.0.1:50053)")
	flag.Parse()

	var peers []string
	if *peersStr != "" {
		peers = strings.Split(*peersStr, ",")
	}

	// Channel to apply committed entries to the state machine
	applyCh := make(chan string)

	// Initialize Raft Node
	node := raft.NewRaftNode(*id, peers, applyCh)

	// Initialize Database Server
	dbServer := server.NewServer(node)

	// Start HTTP Server
	go func() {
		if err := server.StartHTTPServer(*httpAddr, dbServer); err != nil {
			panic(err)
		}
	}()

	// Start gRPC Server (handles both Database and Raft RPCs)
	server.StartGRPCServer(*addr, dbServer)
}
