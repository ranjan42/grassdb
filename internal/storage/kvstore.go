package storage

import (
	"sync"
)

type Store struct {
	mu   sync.RWMutex
	data map[string]string
	wal  *WAL
}

func NewStoreWithWAL(path string) (*Store, error) {
	wal, err := NewWAL(path)
	if err != nil {
		return nil, err
	}

	s := &Store{
		data: make(map[string]string),
		wal:  wal,
	}

	if replayed, err := wal.Replay(); err == nil {
		for k, v := range replayed {
			s.data[k] = v
		}
	}

	return s, nil
}

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.wal.Write(key, value) // In production: handle the error
	s.data[key] = value
}
