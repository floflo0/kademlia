package kademlia

import (
	"fmt"
	"testing"
)

// TestRoutingTableOrder verifies that FindClosestContacts returns contacts ordered by XOR distance
func TestRoutingTableOrder(t *testing.T) {
	myID := NewKademliaID("FFFFFFFF00000000000000000000000000000000000000000000000000000000")
	me := NewContact(myID, "localhost:8000")
	rt := NewRoutingTable(me)

	c1 := NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000000000000000000000000000"), "localhost:8001")
	c2 := NewContact(NewKademliaID("1111111100000000000000000000000000000000000000000000000000000000"), "localhost:8002")
	c3 := NewContact(NewKademliaID("2111111400000000000000000000000000000000000000000000000000000000"), "localhost:8003")

	rt.AddContact(c1)
	rt.AddContact(c2)
	rt.AddContact(c3)

	// Target ID identical to c3
	targetID := NewKademliaID("2111111400000000000000000000000000000000000000000000000000000000")
	contacts := rt.FindClosestContacts(targetID, 20)

	if len(contacts) != 3 {
		t.Fatalf("Expected 3 contacts, got %d", len(contacts))
	}

	// First contact returned MUST be c3 because distance to target is 0
	if !contacts[0].ID.Equals(c3.ID) {
		t.Errorf("Expected closest contact to be %s, got %s", c3.ID.String(), contacts[0].ID.String())
	}
}

// TestRoutingTableDuplicateContact verifies that adding an existing contact does not duplicate it
func TestRoutingTableDuplicateContact(t *testing.T) {
	me := NewContact(NewRandomKademliaID(), "localhost:8000")
	rt := NewRoutingTable(me)

	c1 := NewContact(NewRandomKademliaID(), "localhost:8001")

	// Add the same contact twice
	rt.AddContact(c1)
	rt.AddContact(c1)

	contacts := rt.FindClosestContacts(c1.ID, 20)

	if len(contacts) != 1 {
		t.Errorf("Expected 1 contact after adding duplicate, got %d", len(contacts))
	}
}

// TestBucketSizeLimit verifies that a bucket does not exceed bucketSize (20)
func TestBucketSizeLimit(t *testing.T) {
	me := NewContact(NewKademliaID("0000000000000000000000000000000000000000000000000000000000000000"), "localhost:8000")
	rt := NewRoutingTable(me)

	// Add 25 contacts that fall into the same bucket range
	for i := 0; i < 25; i++ {
		// Generating IDs with similar prefix so they land in nearby/same buckets
		idHex := fmt.Sprintf("80000000000000000000000000000000000000000000000000000000000000%02x", i)
		contact := NewContact(NewKademliaID(idHex), fmt.Sprintf("localhost:80%02d", i))
		rt.AddContact(contact)
	}

	target := NewKademliaID("8000000000000000000000000000000000000000000000000000000000000000")
	bucketIndex := rt.getBucketIndex(target)
	bucketLen := rt.buckets[bucketIndex].Len()

	if bucketLen > bucketSize {
		t.Errorf("Bucket size exceeded limit! Expected <= %d, got %d", bucketSize, bucketLen)
	}
}