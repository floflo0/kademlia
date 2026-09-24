package kademlia

import (
	"container/list"
	"log/slog"
)

const bucketSize = 8

type bucket struct {
	list *list.List
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
	var element *list.Element
	for e := b.list.Front(); e != nil; e = e.Next() {
		nodeID := e.Value.(Contact).ID

		if (contact).ID.Equals(nodeID) {
			element = e
		}
	}

	if element == nil {
		if b.list.Len() < bucketSize {
			b.list.PushFront(contact)
		} else {
			slog.Info("Bucket full", "bucket.list.Len() < bucketSize", b.list.Len() < bucketSize)
			return &contact
		}
	} else {
		b.list.MoveToFront(element)
	}
	return nil
}

// RemoveContact removes the contact with the same ID as contact from the
// bucket. If the contact is not present, removeContact does nothing.
func (b *bucket) RemoveContact(contact Contact) {
	for element := b.list.Front(); element != nil; element = element.Next() {
		nodeID := element.Value.(Contact).ID
		if contact.ID.Equals(nodeID) {
			b.list.Remove(element)
			return
		}
	}
}

// GetContactAndCalcDistance returns an array of Contacts where
// the distance has already been calculated
func (b *bucket) GetContactAndCalcDistance(target *KademliaID) []Contact {
	var contacts []Contact

	for element := b.list.Front(); element != nil; element = element.Next() {
		contact := element.Value.(Contact)
		contact.CalcDistance(target)
		contacts = append(contacts, contact)
	}

	return contacts
}

// Len return the size of the bucket.
func (b *bucket) Len() int {
	return b.list.Len()
}

func (b *bucket) GetContacts() []Contact {
	contacts := make([]Contact, b.list.Len())
	i := 0
	for e := b.list.Front(); e != nil; e = e.Next() {
		contacts[i] = e.Value.(Contact)
		i++
	}
	return contacts
}
