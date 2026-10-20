package kademlia

import (
	"kademlia/internal/core/entities"
	"testing"
)

func createTestContact() *Contact {
	me := NewContact(entities.NewRandomKademliaID(), entities.Address{
		IP:   "127.0.0.1",
		Port: 8000,
	})
	return &me
}

func TestNewContact(t *testing.T) {
	contact := createTestContact()

	if contact == nil {
		t.Fatalf("Expected NewContact to return a non-nil instance")
	}

	if contact.ID == nil {
		t.Errorf("Expected ID to be initialized, got nil")
	}

	str := contact.String()
	if str == "" {
		t.Errorf("Expected non-empty String representation")
	}
}

func TestContactCandidates(t *testing.T) {
	c1 := *createTestContact()
	c2 := *createTestContact()

	target := entities.NewRandomKademliaID()
	c1.CalcDistance(target)
	c2.CalcDistance(target)

	var candidates ContactCandidates
	candidates.Append([]Contact{c1, c2})

	if candidates.Len() != 2 {
		t.Fatalf("expected 2 candidates, got %d", candidates.Len())
	}

	_ = candidates.GetContact(0)
	_ = candidates.GetContacts(2)

	candidates.Swap(0, 1)
	_ = candidates.Less(0, 1)

	candidates.Sort()
	candidates.PopShortList(1)
	if candidates.Len() != 1 {
		t.Fatalf("expected 1 candidate after pop, got %d", candidates.Len())
	}
}
