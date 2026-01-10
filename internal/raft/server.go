package raft

import (
	"context"

	pb "github.com/ranjan42/grassdb/proto"
)

func (rn *RaftNode) RequestVote(ctx context.Context, args *pb.RequestVoteRequest) (*pb.RequestVoteResponse, error) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	// Implement logic:
	// 1. Reply false if term < currentTerm
	// 2. If votedFor is null or candidateId, and candidate's log is at least as up-to-date as receiver's log, grant vote

	if args.Term < int64(rn.currentTerm) {
		return &pb.RequestVoteResponse{Term: int64(rn.currentTerm), VoteGranted: false}, nil
	}

	// If RPC request or response contains term T > currentTerm: set currentTerm = T, convert to follower
	if args.Term > int64(rn.currentTerm) {
		rn.currentTerm = int(args.Term)
		rn.state = Follower
		rn.votedFor = ""
	}

	if rn.votedFor == "" || rn.votedFor == args.CandidateId {
		// TODO: Check log up-to-dateness
		rn.votedFor = args.CandidateId
		rn.resetElectionTimer()
		return &pb.RequestVoteResponse{Term: int64(rn.currentTerm), VoteGranted: true}, nil
	}

	return &pb.RequestVoteResponse{Term: int64(rn.currentTerm), VoteGranted: false}, nil
}

func (rn *RaftNode) AppendEntries(ctx context.Context, args *pb.AppendEntriesRequest) (*pb.AppendEntriesResponse, error) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if args.Term < int64(rn.currentTerm) {
		return &pb.AppendEntriesResponse{Term: int64(rn.currentTerm), Success: false}, nil
	}

	// If RPC request or response contains term T > currentTerm: set currentTerm = T, convert to follower
	if args.Term > int64(rn.currentTerm) {
		rn.currentTerm = int(args.Term)
		rn.state = Follower
		rn.votedFor = ""
	}

	// If we are candidate/leader and receive AppendEntries from valid leader, become follower
	if rn.state != Follower {
		rn.state = Follower
	}

	rn.resetElectionTimer()

	// TODO: Log replication logic

	return &pb.AppendEntriesResponse{Term: int64(rn.currentTerm), Success: true}, nil
}

func (rn *RaftNode) InstallSnapshot(ctx context.Context, args *pb.InstallSnapshotRequest) (*pb.InstallSnapshotResponse, error) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if args.Term > int64(rn.currentTerm) {
		rn.currentTerm = int(args.Term)
		rn.state = Follower
		rn.votedFor = ""
	}

	if args.Term < int64(rn.currentTerm) {
		return &pb.InstallSnapshotResponse{Term: int64(rn.currentTerm)}, nil
	}

	rn.resetElectionTimer()

	// Simplified Snapshot handling:
	// In a full implementation, we would discard the log and replace it with state.
	// For now, we acknowledge the leader's authority.

	return &pb.InstallSnapshotResponse{
		Term: int64(rn.currentTerm),
	}, nil
}
