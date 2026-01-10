package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"grassdb/internal/raft"
	"grassdb/internal/storage"

	pb "github.com/ranjan42/grassdb/proto"

	"google.golang.org/grpc"
)

type DatabaseServer struct {
	pb.UnimplementedDatabaseServer
	store    *storage.Store
	raftNode *raft.RaftNode
}

func NewServer(rn *raft.RaftNode) *DatabaseServer {
	store, err := storage.NewStoreWithWAL(fmt.Sprintf("distdb_%s.wal", rn.ID())) // Use unique WAL per node
	if err != nil {
		log.Fatalf("failed to initialize WAL: %v", err)
	}
	return &DatabaseServer{
		store:    store,
		raftNode: rn,
	}
}

func (s *DatabaseServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	value, found := s.store.Get(req.Key)
	return &pb.GetResponse{Value: value, Found: found}, nil
}

func (s *DatabaseServer) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	// Check if leader
	if !s.raftNode.IsLeader() {
		return &pb.SetResponse{
			Success:  false,
			Error:    "Not Leader",
			LeaderId: s.raftNode.LeaderID(),
		}, nil
	}

	// TODO: Propose to Raft log instead of direct write
	// For now, we write directly to show integration, but real Raft should apply via ApplyCh
	s.store.Set(req.Key, req.Value)
	return &pb.SetResponse{Success: true}, nil
}

func (s *DatabaseServer) RequestVote(ctx context.Context, req *pb.RequestVoteRequest) (*pb.RequestVoteResponse, error) {
	return s.raftNode.RequestVote(ctx, req)
}

func (s *DatabaseServer) AppendEntries(ctx context.Context, req *pb.AppendEntriesRequest) (*pb.AppendEntriesResponse, error) {
	return s.raftNode.AppendEntries(ctx, req)
}

func StartGRPCServer(addr string, rn *raft.RaftNode) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterDatabaseServer(grpcServer, NewServer(rn))

	log.Printf("gRPC server listening on %s", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
