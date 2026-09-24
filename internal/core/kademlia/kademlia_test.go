package kademlia

import (
	"kademlia/internal/adapters"
	"kademlia/internal/core/entities"
	"kademlia/internal/core/ports"
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

	if node.dataStore == nil {
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
