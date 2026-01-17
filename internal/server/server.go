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

	// Attempt to load snapshot
	snapshotPath := fmt.Sprintf("distdb_%s.snap", rn.ID())
	if data, err := storage.LoadSnapshot(snapshotPath); err == nil {
		log.Printf("[%s] Loading snapshot from disk...", rn.ID())
		if err := store.RestoreFromSnapshot(data); err != nil {
			log.Printf("Failed to restore snapshot: %v", err)
		}
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

func (s *DatabaseServer) InstallSnapshot(ctx context.Context, req *pb.InstallSnapshotRequest) (*pb.InstallSnapshotResponse, error) {
	return s.raftNode.InstallSnapshot(ctx, req)
}

func (s *DatabaseServer) TakeSnapshot(ctx context.Context, req *pb.TakeSnapshotRequest) (*pb.TakeSnapshotResponse, error) {
	if !s.raftNode.IsLeader() {
		return &pb.TakeSnapshotResponse{Success: false}, fmt.Errorf("not leader")
	}

	// 1. Get snapshot from store
	data, err := s.store.GetSnapshot()
	if err != nil {
		return &pb.TakeSnapshotResponse{Success: false}, err
	}

	// 2. Pass to RaftNode to persist (using dummy index 1 for POC)
	s.raftNode.Snapshot(1, data)

	return &pb.TakeSnapshotResponse{Success: true}, nil
}

func StartGRPCServer(addr string, srv *DatabaseServer) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterDatabaseServer(grpcServer, srv)

	log.Printf("gRPC server listening on %s", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
