package kademlia

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"kademlia/internal/adapters"
	"kademlia/internal/core/entities"
	"kademlia/internal/core/ports"
	"sync"
	"testing"
)

// Helper function to create a dummy Kademlia node for testing
func createTestNode(port int, id *KademliaID, net ports.Network) *kademlia {
	me := NewContact(id, entities.Address{
		IP:   "127.0.0.1",
		Port: port,
	})
	return NewKademlia(me, net)
}

// TestNewKademlia verifies that a new Kademlia instance is properly initialized
func TestNewKademlia(t *testing.T) {
	node := createTestNode(8000, NewKademliaID("00000000000000000000000000000000000000000000000000000000000000001"), adapters.NewMockNetworkAdapter())

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

func TestKademliaRun(t *testing.T) {
	node := createTestNode(8000, NewKademliaID("00000000000000000000000000000000000000000000000000000000000000001"), adapters.NewMockNetworkAdapter())

	go node.Run()

	// if err != nil {
	// 	t.Fatalf("Expected Run to not return an error")
	// }

	node.Quit()
}

func TestLookupContactFunc(t *testing.T) {
	net := adapters.NewMockNetworkAdapter()
	node1 := createTestNode(8000, NewKademliaID("0000000000000000000000000000000000000000000000000000000000000000"), net)
	node2 := createTestNode(8001, NewKademliaID("0000000000000000000000000000000000000000000000000000000000000028"), net)
	node3 := createTestNode(8002, NewKademliaID("0000000000000000000000000000000000000000000000000000000000000024"), net)
	node4 := createTestNode(8003, NewKademliaID("0000000000000000000000000000000000000000000000000000000000000025"), net)

	node1.RoutingTable.AddContact(node2.me)
	node1.RoutingTable.AddContact(node3.me)
	node2.RoutingTable.AddContact(node1.me)
	node3.RoutingTable.AddContact(node4.me)

	target := NewKademliaID("0000000000000000000000000000000000000000000000000000000000000027")
	go node1.Run()
	go node2.Run()
	go node3.Run()
	go node4.Run()

	candidates, _ := node1.LookupContact(target)

	expectedCandidates := &ContactCandidates{
		[]Contact{node4.me, node3.me, node2.me},
	}

	if len(candidates.contacts) != len(expectedCandidates.contacts) {
		t.Errorf("Expected number of candidates to be %v, got %v", len(expectedCandidates.contacts), len(candidates.contacts))
	}

	for i := range len(expectedCandidates.contacts) {
		if *candidates.contacts[i].ID != *expectedCandidates.contacts[i].ID {
			t.Errorf("Expected candidate ID to be %v, got %v", *expectedCandidates.contacts[i].ID, *candidates.contacts[i].ID)
		}
		if candidates.contacts[i].Address != expectedCandidates.contacts[i].Address {
			t.Errorf("Expected candidate address to be %v, got %v", expectedCandidates.contacts[i].Address, candidates.contacts[i].Address)
		}
	}

	node1.Quit()
	node2.Quit()
	node3.Quit()
	node4.Quit()

}

// TestStoreAndLookupDataLocal verifies storing data locally and retrieving it
func TestStoreAndLookupDataLocal(t *testing.T) {
	node := createTestNode(8000, NewKademliaID("00000000000000000000000000000000000000000000000000000000000000001"), adapters.NewMockNetworkAdapter())
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
	node := createTestNode(8000, NewKademliaID("00000000000000000000000000000000000000000000000000000000000000001"), adapters.NewMockNetworkAdapter())

	// Add a dummy contact to the routing table so LookupContact returns candidates
	dummyContact := NewContact(NewRandomKademliaID(), entities.Address{
		IP:   "127.0.0.1",
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
	node := createTestNode(8000, NewKademliaID("00000000000000000000000000000000000000000000000000000000000000001"), adapters.NewMockNetworkAdapter())
	var wg sync.WaitGroup

	numGoroutines := 50

	// Concurrent writes
	for i := range numGoroutines {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			data := fmt.Appendf(nil, "data-chunk-%d", val)
			node.Store(data)
		}(i)
	}

	// Concurrent reads
	for i := range numGoroutines {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			data := fmt.Appendf(nil, "data-chunk-%d", val)
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
