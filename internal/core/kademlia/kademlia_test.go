package kademlia

import (
	"kademlia/internal/adapters"
	"kademlia/internal/core/entities"
	"kademlia/internal/core/ports"
	"log/slog"
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

// Test the joining procedure
func TestJoinProcedure(t *testing.T) {
	net := adapters.NewMockNetworkAdapter()
	node01 := createTestNode(7997, NewKademliaID("584f29d78cfc54f4d50d39206e6179aaf6a3ba94cf196cd46e982184dba7500e"), net)
	node02 := createTestNode(7998, NewKademliaID("865a5e7a3dff6a43f9a6d5891408dc9b633a776374b78ca869e82b5915699788"), net)
	node0 := createTestNode(7999, NewKademliaID("c2ca635f8aaa10cf452b46f8b76f87e808f16bc386d1dadffb0738c6013412560"), net)
	node1 := createTestNode(8000, NewKademliaID("82d3b0c2c9d99d3c3b4955a37847e263bea9804713457a0524c37ff460aa2387"), net)
	node2 := createTestNode(8001, NewKademliaID("04d7d678f7da903f3fda66d0571820524d776048d100e5281398478f19800a18"), net)
	node3 := createTestNode(8002, NewKademliaID("68cb53c968df317fba321af9ea0328edd371b64087528316e1eccdde48712a69"), net)
	node4 := createTestNode(8003, NewKademliaID("2255b708835f6f174f040e0cf049a7717874e176d27d621fa9430b3efb611aa3"), net)

	node01.RoutingTable.AddContact(node0.me)
	node01.RoutingTable.AddContact(node02.me)
	node02.RoutingTable.AddContact(node01.me)
	node0.RoutingTable.AddContact(node1.me)
	node0.RoutingTable.AddContact(node01.me)
	node1.RoutingTable.AddContact(node0.me)
	node1.RoutingTable.AddContact(node2.me)
	node1.RoutingTable.AddContact(node3.me)
	node2.RoutingTable.AddContact(node1.me)
	node3.RoutingTable.AddContact(node1.me)
	node3.RoutingTable.AddContact(node2.me)

	expectedRoutingTable := NewRoutingTable(node4.me)
	expectedRoutingTable.AddContact(node1.me)
	expectedRoutingTable.AddContact(node3.me)
	expectedRoutingTable.AddContact(node2.me)

	expectedClosest := expectedRoutingTable.FindClosestContacts(node4.me.ID, 3)

	slog.Info("First contact", "node4.firstContact", node4.firstContact)

	go node01.Run(nil)
	go node02.Run(nil)
	go node0.Run(nil)
	go node1.Run(nil)
	go node2.Run(nil)
	go node3.Run(nil)
	go node4.Run(&entities.Address{IP: "127.0.0.1", Port: 8002})

	node4.mux.RLock()
	cond := len(node4.RoutingTable.FindClosestContacts(node4.me.ID, 3)) < 3
	node4.mux.RUnlock()
	for cond {
		node4.mux.Lock()
		cond = len(node4.RoutingTable.FindClosestContacts(node4.me.ID, 3)) < 3
		node4.mux.Unlock()
	}

	node4.mux.RLock()
	testedRoutingTable := node4.RoutingTable.FindClosestContacts(node4.me.ID, 3)
	node4.mux.RUnlock()
	slog.Info("tested table", "testedRoutingTable", testedRoutingTable)

	for i := range len(expectedClosest) {
		if *testedRoutingTable[i].ID != *expectedClosest[i].ID {
			t.Errorf("Expected candidate ID to be %v, got %v", *expectedClosest[i].ID, *testedRoutingTable[i].ID)
		}
	}

	node01.Quit()
	node02.Quit()
	node0.Quit()
	node1.Quit()
	node2.Quit()
	node3.Quit()
	node4.Quit()
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

	go node.Run(nil)

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
	go node1.Run(nil)
	go node2.Run(nil)
	go node3.Run(nil)
	go node4.Run(nil)

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
