package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
)

var (
	ErrHashMismatch = errors.New("The key doesn't match to the the value's hash SHA-256")
	ErrKeyNotFound  = errors.New("Key not found in local storage")
)

// local storage of the node's memory
type DataStore struct {
	mu   sync.RWMutex
	data map[string]string
}

// NewDataStore initializes and returns a new instance of the storage
func NewDataStore() *DataStore {
	return &DataStore{
		data: make(map[string]string),
	}
}

// Put validates that the provided key matches the SHA-256 hash of the value.
// If valid, it stores the key-value pair safely in memory.
// Where:
//   - key: The expected 64-character hex string of the value's SHA-256 hash.
//   - value: The raw binary data (blob) to store.
func (ds *DataStore) Put(key string, value string) error {

	hash := sha256.Sum256([]byte(value))
	expectedKey := hex.EncodeToString(hash[:])

	// validate Kademlia's requirement
	if key != expectedKey {
		return ErrHashMismatch
	}

	// block access to another threads while we write
	ds.mu.Lock()
	defer ds.mu.Unlock()

	// save the value in map
	ds.data[key] = value
	return nil
}

// Get retrieves the value associated with the given key from the local data store.
// Where:
//   - key: The 64-character hex string key to look up.
//
// Returns the binary blob ([]byte) if found, or ErrKeyNotFound if the key does not exist.
func (ds *DataStore) Get(key string) (string, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	val, exists := ds.data[key]
	if !exists {
		return "", ErrKeyNotFound
	}

	return val, nil
}

// Keys returns a slice containing all stored keys in the local data store.
// This is useful for inspection commands like "show ds".
func (ds *DataStore) Keys() []string {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	keys := make([]string, 0, len(ds.data))
	for k := range ds.data {
		keys = append(keys, k)
	}

	return keys
}
