package server

import (
	"context"
	"log"
	"net"

	"grassdb/internal/storage"
	pb "grassdb/proto"

	"google.golang.org/grpc"
)

type DatabaseServer struct {
	pb.UnimplementedDatabaseServer
	store *storage.Store
}

// func NewServer() *DatabaseServer {
// 	return &DatabaseServer{
// 		store: storage.NewStore(),
// 	}
// }

func NewServer() *DatabaseServer {
	store, err := storage.NewStoreWithWAL("distdb.wal")
	if err != nil {
		log.Fatalf("failed to initialize WAL: %v", err)
	}
	return &DatabaseServer{store: store}
}

func (s *DatabaseServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	value, found := s.store.Get(req.Key)
	return &pb.GetResponse{Value: value, Found: found}, nil
}

func (s *DatabaseServer) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	s.store.Set(req.Key, req.Value)
	return &pb.SetResponse{Success: true}, nil
}

func StartGRPCServer(addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterDatabaseServer(grpcServer, NewServer())

	log.Printf("gRPC server listening on %s", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
