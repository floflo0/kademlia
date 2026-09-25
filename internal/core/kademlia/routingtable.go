package kademlia

import "sync"

const bucketSize = 20
const numberBuckets = IDLength * 8

// RoutingTable definition
// keeps a refrence contact of me and an array of buckets
type RoutingTable struct {
	me      Contact
	mux     sync.RWMutex
	buckets [numberBuckets]*bucket
}

// NewRoutingTable returns a new instance of a RoutingTable.
func NewRoutingTable(me Contact) *RoutingTable {
	routingTable := &RoutingTable{
		me: me,
	}
	for i := range numberBuckets {
		routingTable.buckets[i] = newBucket()
	}
	return routingTable
}

// AddContact add a new contact to the correct Bucket
func (r *RoutingTable) AddContact(contact Contact) {
	bucketIndex := r.getBucketIndex(contact.ID)
	r.mux.RLock()
	bucket := r.buckets[bucketIndex]
	r.mux.RUnlock()
	bucket.AddContact(contact)
}

// RemoveContact remove a contact from his bucket
func (r *RoutingTable) RemoveContact(contact Contact) {
	bucketIndex := r.getBucketIndex(contact.ID)
	r.mux.RLock()
	bucket := r.buckets[bucketIndex]
	r.mux.RUnlock()
	bucket.RemoveContact(contact)
}

// FindClosestContacts finds the count closest Contacts to the target in the RoutingTable
func (r *RoutingTable) FindClosestContacts(target *KademliaID, count int) []Contact {
	var candidates ContactCandidates
	bucketIndex := r.getBucketIndex(target)
	r.mux.RLock()
	bucket := r.buckets[bucketIndex]
	r.mux.RUnlock()

	candidates.Append(bucket.GetContactAndCalcDistance(target))

	for i := 1; (bucketIndex-i >= 0 || bucketIndex+i < IDLength*8) && candidates.Len() < count; i++ {
		if bucketIndex-i >= 0 {
			r.mux.RLock()
			bucket = r.buckets[bucketIndex-i]
			r.mux.RUnlock()
			candidates.Append(bucket.GetContactAndCalcDistance(target))
		}
		if bucketIndex+i < numberBuckets {
			r.mux.RLock()
			bucket = r.buckets[bucketIndex+i]
			r.mux.RUnlock()
			candidates.Append(bucket.GetContactAndCalcDistance(target))
		}
	}

	candidates.Sort()

	if count > candidates.Len() {
		count = candidates.Len()
	}

	return candidates.GetContacts(count)
}

// getBucketIndex get the correct Bucket index for the KademliaID
func (r *RoutingTable) getBucketIndex(id *KademliaID) int {
	r.mux.RLock()
	meId := r.me.ID
	r.mux.RUnlock()
	distance := id.CalcDistance(meId)
	for i := range IDLength {
		for j := range 8 {
			if (distance[i]>>uint8(7-j))&0x1 != 0 {
				return i*8 + j
			}
		}
	}
	return numberBuckets - 1
}

func (t *RoutingTable) getBuckets() []*bucket {
	return t.buckets[:]
}
