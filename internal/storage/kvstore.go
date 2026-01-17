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

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	return val, ok
}

// GetSnapshot returns the serialized state of the store.
func (s *Store) GetSnapshot() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return SerializeStore(s.data)
}

// RestoreFromSnapshot replaces the current state with the snapshot.
func (s *Store) RestoreFromSnapshot(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	newState, err := DeserializeStore(data)
	if err != nil {
		return err
	}
	s.data = newState
	return nil
}
