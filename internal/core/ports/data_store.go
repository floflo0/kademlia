package ports

import (
	"errors"
	"kademlia/internal/core/entities"
)

var (
	ErrHashMismatch = errors.New(
		"key does not match the SHA-256 hash of the value",
	)
	ErrKeyNotFound = errors.New("key not found in the data store")
)

type DataStore interface {
	// Get retrieves the value associated with the given key.
	// It returns ErrKeyNotFound if the key does not exist in the data store.
	Get(key entities.KademliaID) (string, error)

	// Put stores the value under the given key.
	// It returns ErrHashMismatch if the key does not match the SHA-256 hash of
	// the value.
	Put(key entities.KademliaID, value string) error

	// Keys returns all keys currently stored in the data store.
	Keys() []entities.KademliaID
}
