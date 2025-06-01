package storage

import (
	"sync"
)

// type Store struct {
// 	mu   sync.RWMutex
// 	data map[string]string
// }

// func NewStore() *Store {
// 	return &Store{
// 		data: make(map[string]string),
// 	}
// }

// func (s *Store) Get(key string) (string, bool) {
// 	s.mu.RLock()
// 	defer s.mu.RUnlock()
// 	val, ok := s.data[key]
// 	return val, ok
// }

// func (s *Store) Set(key, value string) {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()
// 	s.data[key] = value
// }

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
