package kademlia

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"kademlia/internal/adapters"
	"kademlia/internal/core/ports"
	"sync"
	"testing"
)

// Helper function to create a dummy Kademlia node for testing
func createTestNode() *kademlia {
	me := NewContact(NewRandomKademliaID(), ports.Address{
		IP: "127.0.0.1",
		Port: 8000,
	})
	net := adapters.NewMockNetworkAdapter()
	return NewKademlia(me, net)
}

// TestNewKademlia verifies that a new Kademlia instance is properly initialized
func TestNewKademlia(t *testing.T) {
	node := createTestNode()

	if node == nil {
		t.Fatalf("Expected NewKademlia to return a non-nil instance")
	}

	if node.DataStore == nil {
		t.Errorf("Expected DataStore map to be initialized, got nil")
	}

	if node.RoutingTable == nil {
		t.Errorf("Expected RoutingTable to be initialized, got nil")
	}
}

// TestStoreAndLookupDataLocal verifies storing data locally and retrieving it
func TestStoreAndLookupDataLocal(t *testing.T) {
	node := createTestNode()
	testData := []byte("hello kademlia")

	// 1. Calculate expected hash
	expectedHashBytes := sha256.Sum256(testData)
	expectedHashHex := hex.EncodeToString(expectedHashBytes[:])

	// 2. Store the data
	hashHex := node.Store(testData)

	if hashHex != expectedHashHex {
		t.Errorf("Expected generated hash to be %s, got %s", expectedHashHex, hashHex)
	}

	// 3. Lookup the data locally
	data, _, found := node.LookupData(hashHex)

	if !found {
		t.Errorf("Expected data to be found in local DataStore")
	}

	if !bytes.Equal(data, testData) {
		t.Errorf("Expected retrieved data to be '%s', got '%s'", string(testData), string(data))
	}
}

// TestLookupDataNotFound verifies behavior when requested key does not exist locally
func TestLookupDataNotFound(t *testing.T) {
	node := createTestNode()

	// Add a dummy contact to the routing table so LookupContact returns candidates
	dummyContact := NewContact(NewRandomKademliaID(), ports.Address{
		IP: "127.0.0.1",
		Port: 8001,
	})
	node.RoutingTable.AddContact(dummyContact)

	hashBytes := sha256.Sum256([]byte("non-existent"))
	nonExistentHash := hex.EncodeToString(hashBytes[:])

	data, contacts, found := node.LookupData(nonExistentHash)

	if found {
		t.Errorf("Expected found to be false for non-existent key")
	}

	if data != nil {
		t.Errorf("Expected data to be nil when not found locally, got %v", data)
	}

	// Closest contacts should be returned from routing table
	if len(contacts) == 0 {
		t.Errorf("Expected closest contacts to be returned when key is not found")
	}
}

// TestConcurrentStoreAndLookup tests thread safety under concurrent reads and writes
func TestConcurrentStoreAndLookup(t *testing.T) {
	node := createTestNode()
	var wg sync.WaitGroup

	numGoroutines := 50

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			data := []byte(fmt.Sprintf("data-chunk-%d", val))
			node.Store(data)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			data := []byte(fmt.Sprintf("data-chunk-%d", val))
			hashBytes := sha256.Sum256(data)
			hash := hex.EncodeToString(hashBytes[:])
			node.LookupData(hash)
		}(i)
	}

	wg.Wait()

	// Verify total items stored
	node.mux.RLock()
	storeSize := len(node.DataStore)
	node.mux.RUnlock()

	if storeSize != numGoroutines {
		t.Errorf("Expected DataStore size to be %d after concurrent writes, got %d", numGoroutines, storeSize)
	}
}
