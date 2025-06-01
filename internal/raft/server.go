package raft

import (
	"context"
)

type AppendEntriesArgs struct {
	Term     int
	LeaderID string
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

type RequestVoteArgs struct {
	Term         int
	CandidateID  string
	LastLogIndex int
	LastLogTerm  int
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

func (rn *RaftNode) RequestVote(ctx context.Context, args *RequestVoteArgs) (*RequestVoteReply, error) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if args.Term < rn.currentTerm {
		return &RequestVoteReply{Term: rn.currentTerm, VoteGranted: false}, nil
	}

	if rn.votedFor == "" || rn.votedFor == args.CandidateID {
		rn.votedFor = args.CandidateID
		rn.currentTerm = args.Term
		return &RequestVoteReply{Term: rn.currentTerm, VoteGranted: true}, nil
	}

	return &RequestVoteReply{Term: rn.currentTerm, VoteGranted: false}, nil
}

func (rn *RaftNode) AppendEntries(ctx context.Context, args *AppendEntriesArgs) (*AppendEntriesReply, error) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if args.Term < rn.currentTerm {
		return &AppendEntriesReply{Term: rn.currentTerm, Success: false}, nil
	}

	rn.currentTerm = args.Term
	rn.state = Follower
	rn.resetElectionTimer()

	return &AppendEntriesReply{Term: rn.currentTerm, Success: true}, nil
}
