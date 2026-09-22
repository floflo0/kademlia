package kademlia

import (
	"kademlia/internal/core/entities"
	"testing"
)

func createTestContact() *Contact {
	me := NewContact(NewRandomKademliaID(), entities.Address{
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
}
