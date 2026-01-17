package raft

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/ranjan42/grassdb/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (rn *RaftNode) getClient(peer string) (pb.DatabaseClient, error) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if client, ok := rn.peerClients[peer]; ok {
		return client, nil
	}

	// Create new connection
	// Note: We are leaking conn here if we don't store it to Close() later.
	// Ideally we store Conn too, but for now ignoring cleanup.
	conn, err := grpc.NewClient(peer, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("did not connect: %v", err)
	}

	client := pb.NewDatabaseClient(conn)
	rn.peerClients[peer] = client
	return client, nil
}

// sendRequestVote sends a RequestVote RPC to a peer.
func (rn *RaftNode) sendRequestVote(peer string, args *pb.RequestVoteRequest) (*pb.RequestVoteResponse, error) {
	c, err := rn.getClient(peer)
	if err != nil {
		return nil, err
	}

	// Set a reasonable timeout
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	return c.RequestVote(ctx, args)
}

// sendAppendEntries sends an AppendEntries RPC to a peer.
func (rn *RaftNode) sendAppendEntries(peer string, args *pb.AppendEntriesRequest) (*pb.AppendEntriesResponse, error) {
	c, err := rn.getClient(peer)
	if err != nil {
		return nil, err
	}

	// Set a reasonable timeout
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	return c.AppendEntries(ctx, args)
}

// BroadcastAppendEntries sends AppendEntries to all peers.
// This is called by the Leader to send heartbeats or log entries.
func (rn *RaftNode) broadcastAppendEntries() {
	rn.mu.Lock()
	currentTerm := rn.currentTerm
	id := rn.id
	// TODO: Add log info here
	rn.mu.Unlock()

	args := &pb.AppendEntriesRequest{
		Term:     int64(currentTerm),
		LeaderId: id,
		// PrevLogIndex: ...
		// PrevLogTerm: ...
		// Entries: ...
		// LeaderCommit: ...
	}

	for _, peer := range rn.peers {
		// Send in goroutine
		go func(p string) {
			resp, err := rn.sendAppendEntries(p, args)
			if err != nil {
				log.Printf("Failed to send AppendEntries to %s: %v", p, err)
				return
			}
			rn.mu.Lock()
			defer rn.mu.Unlock()

			// If response contains higher term, convert to follower
			if resp.Term > int64(rn.currentTerm) {
				rn.currentTerm = int(resp.Term)
				rn.state = Follower
				rn.votedFor = ""
				rn.resetElectionTimer() // ensure we don't start election immediately
			}
		}(peer)
	}
}
