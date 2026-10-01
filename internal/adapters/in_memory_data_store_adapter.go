package adapters

import (
	"kademlia/internal/core/entities"
	"kademlia/internal/core/ports"
	"log/slog"
	"sync"
)

type inMemoryDataStoreAdapter struct {
	mu   sync.RWMutex
	data map[entities.KademliaID]string
}

// NewInMemoryDataStore initializes and returns a new instance of the storage
func NewInMemoryDataStore() ports.DataStore {
	return &inMemoryDataStoreAdapter{
		data: make(map[entities.KademliaID]string),
	}
}

func (ds *inMemoryDataStoreAdapter) Get(
	key entities.KademliaID,
) (string, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	val, exists := ds.data[key]
	if !exists {
		return "", ports.ErrKeyNotFound
	}

	return val, nil
}

func (ds *inMemoryDataStoreAdapter) Put(
	key entities.KademliaID,
	value string,
) error {
	slog.Info("Put", "key", key.String(), "value", value)
	expectedKey := entities.NewKademliaIDFromString(value)
	if !key.Equals(expectedKey) {
		return ports.ErrHashMismatch
	}

	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.data[key] = value
	return nil
}

func (ds *inMemoryDataStoreAdapter) Keys() []entities.KademliaID {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	keys := make([]entities.KademliaID, 0, len(ds.data))
	for key := range ds.data {
		keys = append(keys, key)
	}
	return keys
}
