package kademlia

import (
	"container/list"
	"kademlia/config"
	"kademlia/internal/core/entities"
	"sync"
)

const BUCKET_SIZE = config.K

type Bucket struct {
	list *list.List
	mux  sync.RWMutex
}

// NewBucket returns a new instance of a bucket.
func NewBucket() *Bucket {
	return &Bucket{
		list: list.New(),
	}
}

// AddContact adds the Contact to the front of the bucket
// or moves it to the front of the bucket if it already existed
// If the bucket is Full returns the pointer to the contact
func (b *Bucket) AddContact(contact Contact) *Contact {
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
		if bucketLen < BUCKET_SIZE {
			b.list.PushFront(contact)
		} else {
			return &contact
		}
	} else {
		b.list.MoveToFront(element)
	}
	return nil
}

// RemoveContact remove the Contact from the bucket
func (b *Bucket) RemoveContact(contact Contact) {
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
func (b *Bucket) GetContactAndCalcDistance(
	target *entities.KademliaID,
) []Contact {
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
func (b *Bucket) Len() int {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.list.Len()
}

func (b *Bucket) GetContacts() []Contact {
	b.mux.RLock()
	defer b.mux.RUnlock()
	contacts := make([]Contact, b.list.Len())
	i := 0
	for e := b.list.Front(); e != nil; e = e.Next() {
		contacts[i] = e.Value.(Contact)
		i++
	}
	return contacts
}
