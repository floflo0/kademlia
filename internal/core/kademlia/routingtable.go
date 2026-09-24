package kademlia

const numberBuckets = IDLength * 8

// RoutingTable definition
// keeps a refrence contact of me and an array of buckets
type RoutingTable struct {
	me      Contact
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
func (t *RoutingTable) AddContact(contact Contact) {
	bucketIndex := t.getBucketIndex(contact.ID)
	bucket := t.buckets[bucketIndex]
	bucket.AddContact(contact)
}

// RemoveContact remove a contact from his bucket
func (t *RoutingTable) RemoveContact(contact Contact) {
	bucketIndex := t.getBucketIndex(contact.ID)
	bucket := t.buckets[bucketIndex]
	bucket.RemoveContact(contact)
}

// FindClosestContacts finds the count closest Contacts to the target in the RoutingTable
func (t *RoutingTable) FindClosestContacts(target *KademliaID, count int) []Contact {
	var candidates ContactCandidates
	bucketIndex := t.getBucketIndex(target)
	bucket := t.buckets[bucketIndex]

	candidates.Append(bucket.GetContactAndCalcDistance(target))

	for i := 1; (bucketIndex-i >= 0 || bucketIndex+i < IDLength*8) && candidates.Len() < count; i++ {
		if bucketIndex-i >= 0 {
			bucket = t.buckets[bucketIndex-i]
			candidates.Append(bucket.GetContactAndCalcDistance(target))
		}
		if bucketIndex+i < numberBuckets {
			bucket = t.buckets[bucketIndex+i]
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
func (t *RoutingTable) getBucketIndex(id *KademliaID) int {
	distance := id.CalcDistance(t.me.ID)
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
