package kademlia

import (
	"kademlia/internal/adapters"
	"kademlia/internal/core/entities"
	"kademlia/internal/core/ports"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Helper function to create a dummy Kademlia node for testing
func createTestNode(
	port int,
	id *entities.KademliaID,
	net ports.Network,
) *kademlia {
	me := NewContact(id, entities.Address{
		IP:   "127.0.0.1",
		Port: port,
	})
	return NewKademlia(me, net)
}

// Test the joining procedure
func TestJoinProcedure(t *testing.T) {
	net := adapters.NewMockNetworkAdapter(0.0)
	id01, _ := entities.NewKademliaID("584f29d78cfc54f4d50d39206e6179aaf6a3ba94cf196cd46e982184dba7500e")
	id02, _ := entities.NewKademliaID("865a5e7a3dff6a43f9a6d5891408dc9b633a776374b78ca869e82b5915699788")
	id0, _ := entities.NewKademliaID("c2ca635f8aaa10cf452b46f8b76f87e808f16bc386d1dadffb0738c601341256")
	id1, _ := entities.NewKademliaID("82d3b0c2c9d99d3c3b4955a37847e263bea9804713457a0524c37ff460aa2387")
	id2, _ := entities.NewKademliaID("04d7d678f7da903f3fda66d0571820524d776048d100e5281398478f19800a18")
	id3, _ := entities.NewKademliaID("68cb53c968df317fba321af9ea0328edd371b64087528316e1eccdde48712a69")
	id4, _ := entities.NewKademliaID("2255b708835f6f174f040e0cf049a7717874e176d27d621fa9430b3efb611aa3")
	node01 := createTestNode(7997, id01, net)
	node02 := createTestNode(7998, id02, net)
	node0 := createTestNode(7999, id0, net)
	node1 := createTestNode(8000, id1, net)
	node2 := createTestNode(8001, id2, net)
	node3 := createTestNode(8002, id3, net)
	node4 := createTestNode(8003, id4, net)

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
	id, _ := entities.NewKademliaID("00000000000000000000000000000000000000000000000000000000000000001")
	node := createTestNode(8000, id, adapters.NewMockNetworkAdapter(0.0))

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
	id, _ := entities.NewKademliaID("00000000000000000000000000000000000000000000000000000000000000001")
	node := createTestNode(8000, id, adapters.NewMockNetworkAdapter(0.0))

	go node.Run(nil)

	// if err != nil {
	// 	t.Fatalf("Expected Run to not return an error")
	// }

	node.Quit()
}

func TestLookupContactFunc(t *testing.T) {
	net := adapters.NewMockNetworkAdapter(0.0)
	id1, _ := entities.NewKademliaID("0000000000000000000000000000000000000000000000000000000000000000")
	id2, _ := entities.NewKademliaID("0000000000000000000000000000000000000000000000000000000000000028")
	id3, _ := entities.NewKademliaID("0000000000000000000000000000000000000000000000000000000000000024")
	id4, _ := entities.NewKademliaID("0000000000000000000000000000000000000000000000000000000000000025")
	node1 := createTestNode(8000, id1, net)
	node2 := createTestNode(8001, id2, net)
	node3 := createTestNode(8002, id3, net)
	node4 := createTestNode(8003, id4, net)

	node1.RoutingTable.AddContact(node2.me)
	node1.RoutingTable.AddContact(node3.me)
	node2.RoutingTable.AddContact(node1.me)
	node3.RoutingTable.AddContact(node4.me)
	node4.RoutingTable.AddContact(node3.me)

	target, _ := entities.NewKademliaID("0000000000000000000000000000000000000000000000000000000000000027")
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

func TestLookupValueFunc(t *testing.T) {
	net := adapters.NewMockNetworkAdapter(0.0)
	id1, _ := entities.NewKademliaID("0000000000000000000000000000000000000000000000000000000000000000")
	id2, _ := entities.NewKademliaID("0000000000000000000000000000000000000000000000000000000000000028")
	id3, _ := entities.NewKademliaID("64ec88ca00b268e5ba1a35678a1b5316d212f4f366b2477232534a8aeca37f00")
	id4, _ := entities.NewKademliaID("64ec88ca00b268e5ba1a35678a1b5316d212f4f366b2477232534a8aeca37f30")
	id5, _ := entities.NewKademliaID("64ec88ca00b268e5ba1a35678a1b5316d212f4f366b2477232534a8aeca37f3c")
	node1 := createTestNode(8000, id1, net)
	node2 := createTestNode(8001, id2, net)
	node3 := createTestNode(8002, id3, net)
	node4 := createTestNode(8003, id4, net)
	node5 := createTestNode(8004, id5, net)

	node1.RoutingTable.AddContact(node2.me)
	node1.RoutingTable.AddContact(node3.me)
	node2.RoutingTable.AddContact(node1.me)
	node3.RoutingTable.AddContact(node4.me)
	node3.RoutingTable.AddContact(node5.me)
	valueTested := "Hello world"
	key := entities.NewKademliaIDFromString(valueTested)
	err := node4.dataStore.Put(*key, valueTested)
	if err != nil {
		t.Fatalf(
			"Put(%q, %q) returns unexpected error: %v",
			key.String(),
			valueTested,
			err,
		)
	}

	go node1.Run(nil)
	go node2.Run(nil)
	go node3.Run(nil)
	go node4.Run(nil)
	go node5.Run(nil)
	time.Sleep(5 * time.Millisecond)

	value, _, candidates, _ := node1.LookupValue(key)

	expectedValue := valueTested

	require.Eventually(t, func() bool {
		found, err := node5.dataStore.Get(*key)
		return err == nil && found != ""
	}, 100*time.Millisecond, 5*time.Millisecond)

	if candidates != nil {
		t.Errorf("Expected number of candidates to be %v, got %v", 0, len(candidates.contacts))
	}

	if *value != expectedValue {
		t.Errorf("Expected value to be %v, got %v", expectedValue, *value)
	}
}

func TestStoreFunc(t *testing.T) {
	net := adapters.NewMockNetworkAdapter(0.0)
	id1, _ := entities.NewKademliaID("0000000000000000000000000000000000000000000000000000000000000000")
	id2, _ := entities.NewKademliaID("0000000000000000000000000000000000000000000000000000000000000028")
	id3, _ := entities.NewKademliaID("64ec88ca00b268e5ba1a35678a1b5316d212f4f366b2477232534a8aeca37f00")
	id4, _ := entities.NewKademliaID("64ec88ca00b268e5ba1a35678a1b5316d212f4f366b2477232534a8aeca37f30")
	id5, _ := entities.NewKademliaID("64ec88ca00b268e5ba1a35678a1b5316d212f4f366b2477232534a8aeca37f3c")
	node1 := createTestNode(8000, id1, net)
	node2 := createTestNode(8001, id2, net)
	node3 := createTestNode(8002, id3, net)
	node4 := createTestNode(8003, id4, net)
	node5 := createTestNode(8004, id5, net)

	node1.RoutingTable.AddContact(node2.me)
	node1.RoutingTable.AddContact(node3.me)
	node2.RoutingTable.AddContact(node1.me)
	node3.RoutingTable.AddContact(node4.me)
	node3.RoutingTable.AddContact(node5.me)
	valueTested := "Hello world"
	key := entities.NewKademliaIDFromString(valueTested)

	go node1.Run(nil)
	go node2.Run(nil)
	go node3.Run(nil)
	go node4.Run(nil)
	go node5.Run(nil)
	time.Sleep(5 * time.Millisecond)

	node1.Store(key, valueTested)

	require.Eventually(t, func() bool {
		found, err := node5.dataStore.Get(*key)
		found1, err := node4.dataStore.Get(*key)
		found2, err := node3.dataStore.Get(*key)
		found3, err := node2.dataStore.Get(*key)
		return err == nil && found != "" && found1 != "" && found2 != "" && found3 != ""
	}, 100*time.Millisecond, 5*time.Millisecond)
}
