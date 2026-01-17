package storage

import (
	"encoding/json"
	"os"
)

// SnapshotMetadata holds info about the snapshot
type SnapshotMetadata struct {
	LastIncludedIndex int
	LastIncludedTerm  int
	// Timestamp, etc.
}

// SaveSnapshot saves the snapshot data to a file.
func SaveSnapshot(path string, data []byte) error {
	// Write atomically: write to temp file then rename
	tmpPath := path + ".tmp"
	err := os.WriteFile(tmpPath, data, 0644)
	if err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

// LoadSnapshot reads the snapshot data from a file.
func LoadSnapshot(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// SerializeStore serializes the KV map to JSON.
func SerializeStore(data map[string]string) ([]byte, error) {
	return json.Marshal(data)
}

// DeserializeStore deserializes JSON to a KV map.
func DeserializeStore(data []byte) (map[string]string, error) {
	var storedData map[string]string
	err := json.Unmarshal(data, &storedData)
	return storedData, err
}
