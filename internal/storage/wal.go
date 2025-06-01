package storage

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

type WAL struct {
	file *os.File
	mu   sync.Mutex
}

func NewWAL(path string) (*WAL, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return &WAL{file: file}, nil
}

func (w *WAL) Write(key, value string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	_, err := fmt.Fprintf(w.file, "%s=%s\n", key, value)
	return err
}

func (w *WAL) Replay() (map[string]string, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	_, err := w.file.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(w.file)
	data := make(map[string]string)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), "=", 2)
		if len(parts) == 2 {
			data[parts[0]] = parts[1]
		}
	}
	return data, scanner.Err()
}

func (w *WAL) Close() error {
	return w.file.Close()
}
