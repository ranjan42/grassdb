package handler

import (
	"encoding/json"
	"fmt"
	"grassdb/internal/storage"
	"net/http"
	"strings"
	"sync"
)

// Global store instance (persists across warm invocations only)
var (
	store    *storage.Store
	initOnce sync.Once
	initErr  error
)

func getStore() (*storage.Store, error) {
	initOnce.Do(func() {
		// Use /tmp for ephemeral storage on Vercel
		s, err := storage.NewStoreWithWAL("/tmp/grassdb_vercel.wal")
		if err != nil {
			initErr = err
			return
		}
		store = s
	})
	return store, initErr
}

// Handler is the Vercel entrypoint
func Handler(w http.ResponseWriter, r *http.Request) {
	// CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	st, err := getStore()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to init store: %v", err), http.StatusInternalServerError)
		return
	}

	// Dispatch based on Path or Method
	// If path contains /get or Method is GET
	// If path contains /set or Method is POST

	// Check path suffix
	path := r.URL.Path

	// Logic:
	// If ends with /get -> handleGet
	// If ends with /set -> handleSet
	// If it's just /api (root), dispatch by method.

	if strings.HasSuffix(path, "/get") || (path == "/api" && r.Method == http.MethodGet) {
		handleGet(w, r, st)
		return
	}

	if strings.HasSuffix(path, "/set") || (path == "/api" && r.Method == http.MethodPost) {
		handleSet(w, r, st)
		return
	}

	// Fallback: Dispatch by method if path is ambiguous
	if r.Method == http.MethodGet {
		handleGet(w, r, st)
	} else if r.Method == http.MethodPost {
		handleSet(w, r, st)
	} else {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGet(w http.ResponseWriter, r *http.Request, st *storage.Store) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "missing key", http.StatusBadRequest)
		return
	}

	val, found := st.Get(key)
	response := map[string]interface{}{
		"value": val,
		"found": found,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleSet(w http.ResponseWriter, r *http.Request, st *storage.Store) {
	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	// Try creating a decoder
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	st.Set(req.Key, req.Value)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
