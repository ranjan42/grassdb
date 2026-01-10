package raft

import (
	"log"
	"math/rand"
	"sync"
	"time"

	pb "github.com/ranjan42/grassdb/proto"
)

type State string

const (
	Follower  State = "Follower"
	Candidate State = "Candidate"
	Leader    State = "Leader"
)

type RaftNode struct {
	mu          sync.Mutex
	id          string
	state       State
	currentTerm int
	votedFor    string
	log         []*pb.LogEntry

	// Volatile state on all servers
	commitIndex int
	lastApplied int

	// Volatile state on leaders
	nextIndex  map[string]int
	matchIndex map[string]int

	peers              []string
	electionTimer      *time.Timer
	heartbeatTimer     *time.Timer
	leaderTimeoutTimer *time.Timer // Timer to detect leader failure
	applyCh            chan string // simplified state machine
	peerClients        map[string]pb.DatabaseClient

	// Snapshot state
	lastIncludedIndex int
	lastIncludedTerm  int
}

func NewRaftNode(id string, peers []string, applyCh chan string) *RaftNode {
	rn := &RaftNode{
		id:                 id,
		peers:              peers,
		state:              Follower,
		applyCh:            applyCh,
		electionTimer:      time.NewTimer(randomElectionTimeout()),
		heartbeatTimer:     time.NewTimer(100 * time.Millisecond),
		leaderTimeoutTimer: time.NewTimer(randomLeaderTimeout()),
		log:                make([]*pb.LogEntry, 0),
		nextIndex:          make(map[string]int),
		matchIndex:         make(map[string]int),
		peerClients:        make(map[string]pb.DatabaseClient),
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
			log.Printf("[%s] Election timeout, becoming Candidate", rn.id)
			rn.mu.Unlock()
			return
		case <-rn.leaderTimeoutTimer.C:
			rn.mu.Lock()
			rn.state = Candidate
			rn.mu.Unlock()
			return
		default:
			rn.mu.Lock()
			if rn.state != Follower {
				rn.mu.Unlock()
				return
			}
			rn.mu.Unlock()
			time.Sleep(10 * time.Millisecond) // IDLE wait
		}
	}
}

func (rn *RaftNode) runCandidate() {
	rn.mu.Lock()
	rn.currentTerm++
	rn.votedFor = rn.id
	rn.resetElectionTimer()
	term := rn.currentTerm
	id := rn.id
	rn.mu.Unlock()

	// Send RequestVote to all peers
	votes := 1 // Vote for self
	voteCh := make(chan bool, len(rn.peers))

	for _, peer := range rn.peers {
		go func(p string) {
			args := &pb.RequestVoteRequest{
				Term:        int64(term),
				CandidateId: id,
				// LastLogIndex: ...
				// LastLogTerm: ...
			}
			resp, err := rn.sendRequestVote(p, args)
			if err != nil {
				voteCh <- false
				return
			}
			voteCh <- resp.VoteGranted
		}(peer)
	}

	// Wait for votes
	for i := 0; i < len(rn.peers); i++ {
		vote := <-voteCh
		if vote {
			votes++
		}
	}

	rn.mu.Lock()
	if rn.state != Candidate {
		rn.mu.Unlock()
		return
	}

	// Majority check (myself + peers)
	if votes > (len(rn.peers)+1)/2 {
		rn.state = Leader
		log.Printf("[%s] Won election! Becoming Leader for term %d", rn.id, rn.currentTerm)
		rn.votedFor = ""
		// Initialize leader state
		for _, p := range rn.peers {
			rn.nextIndex[p] = len(rn.log) // Index of next log entry to send
			rn.matchIndex[p] = -1         // Index of highest log entry known to be replicated
		}
	} else {
		// Failed election, stay candidate (loop will retry or follower)
		// Back to follower effectively if timeout?
		// Actually runCandidate just finishes, main loop calls it again if state is Candidate
		// But we should verify if we stepped down
	}
	rn.mu.Unlock()
}

func (rn *RaftNode) runLeader() {
	rn.resetHeartbeatTimer()
	rn.leaderTimeoutTimer.Stop() // Stop the election timer while leader
	for {
		select {
		case <-rn.heartbeatTimer.C:
			rn.sendHeartbeats()
			rn.resetHeartbeatTimer()
		default:
			rn.mu.Lock()
			if rn.state != Leader {
				rn.mu.Unlock()
				return
			}
			rn.mu.Unlock()
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// ID returns the node's unique identifier.
func (rn *RaftNode) ID() string {
	return rn.id
}

// IsLeader checks if the node is currently the leader.
func (rn *RaftNode) IsLeader() bool {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	return rn.state == Leader
}

// LeaderID returns the current leader's ID if known.
// Note: This is an optimization; it might be stale.
func (rn *RaftNode) LeaderID() string {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	if rn.state == Leader {
		return rn.id
	}
	// TODO: Track current leader ID in follow state
	return "unknown"
}

func (rn *RaftNode) sendHeartbeats() {
	rn.broadcastAppendEntries()
}

func (rn *RaftNode) resetElectionTimer() {
	if !rn.electionTimer.Stop() {
		select {
		case <-rn.electionTimer.C:
		default:
		}
	}
	rn.electionTimer.Reset(randomElectionTimeout())
}

func (rn *RaftNode) resetHeartbeatTimer() {
	if !rn.heartbeatTimer.Stop() {
		select {
		case <-rn.heartbeatTimer.C:
		default:
		}
	}
	rn.heartbeatTimer.Reset(100 * time.Millisecond)
}

func (rn *RaftNode) resetLeaderTimeoutTimer() {
	if !rn.leaderTimeoutTimer.Stop() {
		select {
		case <-rn.leaderTimeoutTimer.C:
		default:
		}
	}
	rn.leaderTimeoutTimer.Reset(randomLeaderTimeout())
}

func randomElectionTimeout() time.Duration {
	return time.Duration(300+rand.Intn(300)) * time.Millisecond
}

func randomLeaderTimeout() time.Duration {
	return time.Duration(600+rand.Intn(400)) * time.Millisecond
}

// Snapshot creates a snapshot of the state machine up to the given index.
// In this simplified version, it truncates the log.
func (rn *RaftNode) Snapshot(index int, data []byte) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if index <= rn.lastIncludedIndex {
		return // Already snapshotted
	}

	// Calculate actual log index
	// Logical index = lastIncludedIndex + len(log) if we kept everything
	// We need to find the entry at 'index'

	// For now, let's assume index is valid and in our current log.
	// realIndex := index - rn.lastIncludedIndex

	// We basically just clear the log for now as a POC of compaction.
	// In a real impl, we keep the last entry as the new "prevLog" for AppendEntries consistency.

	rn.lastIncludedIndex = index
	rn.lastIncludedTerm = rn.currentTerm // Simplified

	// Truncate log (keep nothing for now, or maybe keep tails)
	rn.log = make([]*pb.LogEntry, 0)

	log.Printf("[%s] Created snapshot at index %d", rn.id, index)

	// TODO: Write 'data' (snapshot state) to disk
}
