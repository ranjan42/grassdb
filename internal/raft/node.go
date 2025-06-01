package raft

type PersistentState struct {
	CurrentTerm int
	VotedFor    string
}

type VolatileState struct {
	CommitIndex int
	LastApplied int
}

type NodeState struct {
	Persistent PersistentState
	Volatile   VolatileState
	Log        *Log
}
