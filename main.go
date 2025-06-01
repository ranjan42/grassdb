// main.go
package main

import (
	"flag"
	"log"
	"net"
	"strings"

	"grassdb/internal/raft"

	pb "github.com/ranjan/grassdb/proto"

	"google.golang.org/grpc"
)

func main() {
	id := flag.String("id", "", "Unique node ID")
	port := flag.String("port", "", "Port to listen on")
	peersStr := flag.String("peers", "", "Comma-separated list of peer addresses (e.g. node2:50052,node3:50053)")
	flag.Parse()

	if *id == "" || *port == "" {
		log.Fatalf("Both --id and --port are required")
	}
	peers := strings.Split(*peersStr, ",")

	applyCh := make(chan string)
	node := raft.NewRaftNode(*id, peers, applyCh)

	lis, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", *port, err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterRaftServer(grpcServer, node)

	log.Printf("Raft node %s listening on port %s\n", *id, *port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC server: %v", err)
	}
}
