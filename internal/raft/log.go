package raft

import (
	"sync"
)

type LogEntry struct {
	Term    int
	Command string
}

type Log struct {
	mu      sync.Mutex
	entries []LogEntry
}

func NewLog() *Log {
	return &Log{
		entries: make([]LogEntry, 0),
	}
}

func (l *Log) Append(entry LogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, entry)
}

func (l *Log) Get(index int) (LogEntry, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if index < 0 || index >= len(l.entries) {
		return LogEntry{}, false
	}
	return l.entries[index], true
}

func (l *Log) LastIndex() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.entries) - 1
}

func (l *Log) LastTerm() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.entries) == 0 {
		return 0
	}
	return l.entries[len(l.entries)-1].Term
}
