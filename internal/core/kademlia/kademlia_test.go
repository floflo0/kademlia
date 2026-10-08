package kademlia

import (
	"fmt"
	"kademlia/internal/adapters"
	"kademlia/internal/core/entities"
	"kademlia/internal/core/ports"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Helper function to create a dummy Kademlia node for testing
func createTestNode(
	port int,
	net ports.Network,
) *kademlia {
	return NewKademlia(entities.Address{
		IP:   "127.0.0.1",
		Port: port,
	}, net)
}

// Test the joining procedure
func TestJoinProcedure(t *testing.T) {
	net := adapters.NewMockNetworkAdapter(0.0)
	node01 := createTestNode(7997, net)
	node02 := createTestNode(7998, net)
	node0 := createTestNode(7999, net)
	node1 := createTestNode(8000, net)
	node2 := createTestNode(8001, net)
	node3 := createTestNode(8002, net)
	node4 := createTestNode(8003, net)

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

	go node01.Run(nil)
	go node02.Run(nil)
	go node0.Run(nil)
	go node1.Run(nil)
	go node2.Run(nil)
	go node3.Run(nil)
	time.Sleep(100 * time.Millisecond)
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
			t.Fatalf(
				"Expected candidate ID to be %q, got %q",
				expectedClosest[i].ID.String(),
				testedRoutingTable[i].ID.String(),
			)
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
	node := createTestNode(8000, adapters.NewMockNetworkAdapter(0.0))

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
	node := createTestNode(8000, adapters.NewMockNetworkAdapter(0.0))
	errChan := make(chan error, 1)

	go func() {
		errChan <- node.Run(nil)
	}()
	time.Sleep(5 * time.Millisecond)

	err := node.Quit()

	if err != nil {
		t.Fatalf("Quit() returns unexpected error: %v", err)
	}

	err = <-errChan
	if err != nil {
		t.Fatalf("Run(nil) returns unexpected error: %v", err)
	}
}

func TestLookupContactFunc(t *testing.T) {
	net := adapters.NewMockNetworkAdapter(0.0)
	node1 := createTestNode(8000, net)
	node2 := createTestNode(8001, net)
	node3 := createTestNode(8002, net)
	node4 := createTestNode(8003, net)

	node1.RoutingTable.AddContact(node2.me)
	node1.RoutingTable.AddContact(node3.me)
	node2.RoutingTable.AddContact(node1.me)
	node3.RoutingTable.AddContact(node4.me)
	node4.RoutingTable.AddContact(node3.me)

	target, _ := entities.NewKademliaID(
		"0000000000000000000000000000000000000000000000000000000000000001",
	)
	go node1.Run(nil)
	go node2.Run(nil)
	go node3.Run(nil)
	go node4.Run(nil)
	time.Sleep(5 * time.Millisecond)

	candidates, _ := node1.LookupContact(target)

	expectedCandidates := &ContactCandidates{
		[]Contact{node2.me, node4.me, node3.me},
	}
	slog.Info("foo", "bar", candidates.contacts)

	if len(candidates.contacts) != len(expectedCandidates.contacts) {
		t.Fatalf("Expected number of candidates to be %v, got %v", len(expectedCandidates.contacts), len(candidates.contacts))
	}

	for i := range len(expectedCandidates.contacts) {
		if *candidates.contacts[i].ID != *expectedCandidates.contacts[i].ID {
			t.Fatalf("Expected candidate ID to be %v, got %v", *expectedCandidates.contacts[i].ID, *candidates.contacts[i].ID)
		}
		if candidates.contacts[i].Address != expectedCandidates.contacts[i].Address {
			t.Fatalf("Expected candidate address to be %v, got %v", expectedCandidates.contacts[i].Address, candidates.contacts[i].Address)
		}
	}

	node1.Quit()
	node2.Quit()
	node3.Quit()
	node4.Quit()

}

func TestLookupValueFunc(t *testing.T) {
	net := adapters.NewMockNetworkAdapter(0.0)
	node1 := createTestNode(8000, net)
	node2 := createTestNode(8001, net)
	node3 := createTestNode(8002, net)
	node4 := createTestNode(8003, net)
	node5 := createTestNode(8004, net)

	node1.RoutingTable.AddContact(node2.me)
	node1.RoutingTable.AddContact(node3.me)
	node2.RoutingTable.AddContact(node1.me)
	node3.RoutingTable.AddContact(node4.me)
	node3.RoutingTable.AddContact(node5.me)
	valueTested := "7"
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
		_, err := node5.dataStore.Get(*key)
		return err == nil
	}, 100*time.Millisecond, 5*time.Millisecond)

	if candidates != nil {
		t.Fatalf("Expected number of candidates to be %v, got %v", 0, len(candidates.contacts))
	}

	if *value != expectedValue {
		t.Fatalf("Expected value to be %v, got %v", expectedValue, *value)
	}
}

func TestStoreFunc(t *testing.T) {
	net := adapters.NewMockNetworkAdapter(0.0)
	node1 := createTestNode(8000, net)
	node2 := createTestNode(8001, net)
	node3 := createTestNode(8002, net)
	node4 := createTestNode(8003, net)
	node5 := createTestNode(8004, net)

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

	err := node1.Store(key, valueTested)
	if err != nil {
		t.Fatalf(
			"Store(%q, %q) returns unexpected error: %v",
			key.String(),
			valueTested,
			err,
		)
	}

	require.Eventually(t, func() bool {
		_, err1 := node5.dataStore.Get(*key)
		_, err2 := node4.dataStore.Get(*key)
		_, err3 := node3.dataStore.Get(*key)
		_, err4 := node2.dataStore.Get(*key)
		return err1 == nil && err2 == nil && err3 == nil && err4 == nil
	}, 100*time.Millisecond, 5*time.Millisecond)
}

func Test1000Nodes(t *testing.T) {
	network := adapters.NewMockNetworkAdapter(0.0)
	numberNodes := 1000
	nodes := make([]Kademlia, numberNodes)

	for i := range numberNodes {
		nodes[i] = NewKademlia(entities.Address{
			IP:   "127.0.0.1",
			Port: 8000 + i,
		}, network)
	}

	go nodes[0].Run(nil)
	time.Sleep(100 * time.Millisecond)

	for i := 1; i < numberNodes; i++ {
		go nodes[i].Run(&entities.Address{
			IP:   "127.0.0.1",
			Port: 8000,
		})
	}
	t.Cleanup(func() {
		for i := range numberNodes {
			err := nodes[i].Quit()
			if err != nil {
				t.Errorf("Quit() returns unexpected error: %v", err)
			}
		}
	})

	timeout := time.After(20 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		var ready atomic.Bool
		ready.Store(true)

		var waitGroup sync.WaitGroup
		for i := range numberNodes {
			waitGroup.Go(func() {
				address := entities.Address{
					IP:   "127.0.0.1",
					Port: 8000 + i,
				}
				_, err := nodes[0].Ping(address)
				if err != nil {
					ready.Store(false)
				}
			})
		}
		waitGroup.Wait()

		if ready.Load() {
			break
		}

		select {
		case <-timeout:
			t.Fatal("timed out waiting for all nodes to start")
		case <-ticker.C:
		}
	}

	keys := make([]entities.KademliaID, numberNodes)
	if !t.Run("put", func(t *testing.T) {
		var waitGroup sync.WaitGroup
		for i := range numberNodes {
			waitGroup.Go(func() {
				data := fmt.Sprintf("data %d", i)
				key, err := nodes[i].Put(data)
				if err != nil {
					t.Errorf("Put(%q) returns unexpected error: %v", data, err)
					return
				}
				keys[i] = *key
			})
		}
		waitGroup.Wait()
	}) {
		return
	}

	if !t.Run("get", func(t *testing.T) {
		var waitGroup sync.WaitGroup
		for i := range numberNodes {
			waitGroup.Go(func() {
				j := numberNodes - 1 - i
				key := keys[j].String()
				data, _, err := nodes[i].GetValue(key)
				if err != nil {
					t.Errorf("Get(%q) returns unexpected error: %v", key, err)
				}
				expectedData := fmt.Sprintf("data %d", j)
				if data == nil {
					t.Errorf(
						"Get(%q) returned nil data; want %q",
						key,
						expectedData,
					)
					return
				}
				if *data != expectedData {
					t.Errorf(
						"Get(%q) data = %q; want %q",
						key,
						*data,
						expectedData,
					)
				}
			})
		}
		waitGroup.Wait()
	}) {
		return
	}
}
