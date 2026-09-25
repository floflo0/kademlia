package kademlia

import (
	"container/list"
	"log/slog"
	"sync"
)

const bucketSize = 8

type bucket struct {
	list *list.List
	mux  sync.RWMutex
}

// newBucket returns a new instance of a bucket.
func newBucket() *bucket {
	return &bucket{
		list: list.New(),
	}
}

// AddContact adds the Contact to the front of the bucket
// or moves it to the front of the bucket if it already existed
// If the bucket is Full returns the pointer to the contact
func (b *bucket) AddContact(contact Contact) *Contact {
	b.mux.Lock()
	defer b.mux.Unlock()
	var element *list.Element

	for e := b.list.Front(); e != nil; e = e.Next() {
		nodeID := e.Value.(Contact).ID

		if (contact).ID.Equals(nodeID) {
			element = e
		}
	}

	if element == nil {
		bucketLen := b.list.Len()
		if bucketLen < bucketSize {
			b.list.PushFront(contact)
		} else {
			slog.Info("Bucket full", "bucket.list.Len() < bucketSize", bucketLen < bucketSize)
			return &contact
		}
	} else {
		b.list.MoveToFront(element)
	}
	return nil
}

// RemoveContact remove the Contact from the bucket
func (b *bucket) RemoveContact(contact Contact) {
	b.mux.Lock()
	defer b.mux.Unlock()
	var element *list.Element

	for e := b.list.Front(); e != nil; e = e.Next() {
		nodeID := e.Value.(Contact).ID

		if (contact).ID.Equals(nodeID) {
			element = e
		}
	}

	if element != nil {
		b.list.Remove(element)
	}
}

// GetContactAndCalcDistance returns an array of Contacts where
// the distance has already been calculated
func (b *bucket) GetContactAndCalcDistance(target *KademliaID) []Contact {
	var contacts []Contact
	b.mux.RLock()
	for elt := b.list.Front(); elt != nil; elt = elt.Next() {
		contact := elt.Value.(Contact)
		contact.CalcDistance(target)
		contacts = append(contacts, contact)
	}
	b.mux.RUnlock()
	return contacts
}

// Len return the size of the bucket
func (b *bucket) Len() int {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.list.Len()
}

func (b *bucket) GetContacts() []Contact {
	b.mux.RLock()
	contacts := make([]Contact, b.list.Len())
	i := 0
	for e := b.list.Front(); e != nil; e = e.Next() {
		contacts[i] = e.Value.(Contact)
		i++
	}
	b.mux.RUnlock()
	return contacts
}
