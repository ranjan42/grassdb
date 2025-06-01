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
	mu                 sync.Mutex
	id                 string
	state              State
	currentTerm        int
	votedFor           string
	peers              []string
	electionTimer      *time.Timer
	heartbeatTimer     *time.Timer
	leaderTimeoutTimer *time.Timer // Timer to detect leader failure
	applyCh            chan string // simplified state machine
}

func NewRaftNode(id string, peers []string, applyCh chan string) *RaftNode {
	rn := &RaftNode{
		id:                 id,
		peers:              peers,
		state:              Follower,
		applyCh:            applyCh,
		electionTimer:      time.NewTimer(randomElectionTimeout()),
		heartbeatTimer:     time.NewTimer(50 * time.Millisecond),
		leaderTimeoutTimer: time.NewTimer(randomLeaderTimeout()),
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
	rn.resetLeaderTimeoutTimer()
	for {
		select {
		case <-rn.electionTimer.C:
			rn.mu.Lock()
			rn.state = Candidate
			rn.mu.Unlock()
			return
		case <-rn.leaderTimeoutTimer.C:
			rn.mu.Lock()
			rn.state = Candidate
			rn.mu.Unlock()
			return
		default:
			if rn.state != Follower {
				return
			}
		}
	}
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
	rn.resetLeaderTimeoutTimer()
	for {
		select {
		case <-rn.heartbeatTimer.C:
			rn.sendHeartbeats()
			rn.resetHeartbeatTimer()
		case <-rn.leaderTimeoutTimer.C:
			rn.mu.Lock()
			rn.state = Candidate
			rn.mu.Unlock()
			return
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

func (rn *RaftNode) resetLeaderTimeoutTimer() {
	if !rn.leaderTimeoutTimer.Stop() {
		<-rn.leaderTimeoutTimer.C
	}
	rn.leaderTimeoutTimer.Reset(randomLeaderTimeout())
}

func randomElectionTimeout() time.Duration {
	return time.Duration(150+rand.Intn(150)) * time.Millisecond
}

func randomLeaderTimeout() time.Duration {
	return time.Duration(300+rand.Intn(200)) * time.Millisecond
}

// Placeholder function to ensure the package is valid.
func RaftFunction() {}
