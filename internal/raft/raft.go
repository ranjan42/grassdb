package raft

import (
	"math/rand"
	"sync"
	"time"
)

type State string

const (
	Follower  State = "Follower"
	Candidate State = "Candidate"
	Leader    State = "Leader"
)

type RaftNode struct {
	mu             sync.Mutex
	id             string
	state          State
	currentTerm    int
	votedFor       string
	peers          []string
	electionTimer  *time.Timer
	heartbeatTimer *time.Timer
	applyCh        chan string // simplified state machine
}

func NewRaftNode(id string, peers []string, applyCh chan string) *RaftNode {
	rn := &RaftNode{
		id:             id,
		peers:          peers,
		state:          Follower,
		applyCh:        applyCh,
		electionTimer:  time.NewTimer(randomElectionTimeout()),
		heartbeatTimer: time.NewTimer(50 * time.Millisecond),
	}
	go rn.run()
	return rn
}

func (rn *RaftNode) run() {
	for {
		switch rn.state {
		case Follower:
			rn.runFollower()
		case Candidate:
			rn.runCandidate()
		case Leader:
			rn.runLeader()
		}
	}
}

func (rn *RaftNode) runFollower() {
	rn.resetElectionTimer()
	<-rn.electionTimer.C
	rn.mu.Lock()
	rn.state = Candidate
	rn.mu.Unlock()
}

func (rn *RaftNode) runCandidate() {
	rn.mu.Lock()
	rn.currentTerm++
	rn.votedFor = rn.id
	rn.resetElectionTimer()
	rn.mu.Unlock()

	// In a real setup, we would now send RequestVote RPCs to peers

	// Simulate success
	time.Sleep(100 * time.Millisecond)
	rn.mu.Lock()
	rn.state = Leader
	rn.mu.Unlock()
}

func (rn *RaftNode) runLeader() {
	rn.resetHeartbeatTimer()
	for {
		select {
		case <-rn.heartbeatTimer.C:
			rn.sendHeartbeats()
			rn.resetHeartbeatTimer()
		default:
			if rn.state != Leader {
				return
			}
		}
	}
}

func (rn *RaftNode) sendHeartbeats() {
	// In a real system: send AppendEntries RPC with empty entries
	// Here we simulate a heartbeat broadcast
}

func (rn *RaftNode) resetElectionTimer() {
	if !rn.electionTimer.Stop() {
		<-rn.electionTimer.C
	}
	rn.electionTimer.Reset(randomElectionTimeout())
}

func (rn *RaftNode) resetHeartbeatTimer() {
	if !rn.heartbeatTimer.Stop() {
		<-rn.heartbeatTimer.C
	}
	rn.heartbeatTimer.Reset(50 * time.Millisecond)
}

func randomElectionTimeout() time.Duration {
	return time.Duration(150+rand.Intn(150)) * time.Millisecond
}
