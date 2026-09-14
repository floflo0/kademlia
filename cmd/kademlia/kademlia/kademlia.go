package kademlia

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

type Kademlia struct {
	RoutingTable *RoutingTable
	Network      Network
	DataStore    map[string][]byte
	me           Contact
	mux          sync.RWMutex
}

// NewKademlia creates and initializes a new instance of the Kademlia node
func NewKademlia(me Contact, net Network) *Kademlia {
	return &Kademlia{
		RoutingTable: NewRoutingTable(me),
		Network:      net,
		DataStore:    make(map[string][]byte),
		me:           me,
	}
}

// LookupContact does an iterative search of the closest k nodes to a target ID
func (kademlia *Kademlia) LookupContact(target *KademliaID) []Contact {
	// 1. Obtain the initial closest contacts from the local routing table
	candidates := kademlia.RoutingTable.FindClosestContacts(target, 20)

	// TODO: Implement the iterative searching loop (Node Lookup) sending
	// FIND_NODE requests in parallel (alpha=3) to the candidates.

	return candidates
}

// LookupData searches for the value belonging to a key (hash)
// If it finds the value locally, it returns (data, nil, true)
// If it doesn't, it returns (nil, kClosestContacts, false)
func (kademlia *Kademlia) LookupData(hash string) ([]byte, []Contact, bool) {
	kademlia.mux.RLock()
	val, exists := kademlia.DataStore[hash]
	kademlia.mux.RUnlock()

	// If the value is stored locally, return it immediately
	if exists {
		return val, nil, true
	}

	targetID := NewKademliaID(hash)
	closest := kademlia.LookupContact(targetID)

	// TODO: Send RPCs FIND_VALUE to the closest nodes until
	// obtaining the value or running out of contacts.

	return nil, closest, false
}

// Store calculates the SHA-256 key of the data, saves it locally,
// searches for the closest k nodes, and sends a STORE RPC to each one.
func (kademlia *Kademlia) Store(data []byte) string {
	// 1. Calculate K = SHA-256(data) (32 bytes / 64 hex char to match IDLength=32)
	hashBytes := sha256.Sum256(data)
	keyHex := hex.EncodeToString(hashBytes[:])
	keyID := NewKademliaID(keyHex)

	// 2. Save a copy in the local DataStore safely
	kademlia.mux.Lock()
	kademlia.DataStore[keyHex] = data
	kademlia.mux.Unlock()

	// 3. Search for the closest k nodes to the key
	targetNodes := kademlia.LookupContact(keyID)

	// 4. Send a STORE RPC to each of the closest k nodes
	for _, contact := range targetNodes {
		// go kademlia.Network.SendStoreRPC(&contact, keyHex, data)
		_ = contact
	}

	return keyHex
}
