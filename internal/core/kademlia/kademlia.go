package kademlia

import (
	"crypto/sha256"
	"encoding/hex"
	"kademlia/internal/core/ports"
	"log/slog"
	"slices"
	"sync"
)

const k = 4
const alpha = 4
const b = 1

type Kademlia struct {
	RoutingTable *RoutingTable
	Network      ports.Network
	DataStore    map[string][]byte
	me           Contact
	mux          sync.RWMutex
}

// NewKademlia creates and initializes a new instance of the Kademlia node
func NewKademlia(me Contact, net ports.Network) *Kademlia {
	slog.Debug("", "me", me)
	return &Kademlia{
		RoutingTable: NewRoutingTable(me),
		Network:      net,
		DataStore:    make(map[string][]byte),
		me:           me,
	}
}

// LookupContact does an iterative search of the closest k nodes to a target ID
func (kademlia *Kademlia) LookupContact(target *KademliaID) ContactCandidates {
	// 1. Obtain the initial closest contacts from the local routing table
	var candidates ContactCandidates
	var noNewClosest bool
	var probed int
	// TODO: No RPC response reaction,

	var alreadyContacted []Contact

	candidates.Append(kademlia.RoutingTable.FindClosestContacts(target, k))
	candidates.Sort()
	closestNode := candidates.GetContact(0)
	noNewClosest = false

	for (!noNewClosest) && (probed == k) {
		var wg sync.WaitGroup
		ans := make(chan RPCResponse, alpha)

		for nodeCounter := range k {
			wg.Add(1)

			go func() {
				defer wg.Done()
				contact := candidates.GetContact(nodeCounter)
				if slices.Contains(alreadyContacted, contact) == false {
					ansFindNode, err := SendFindNode(kademlia.Network, contact.Address, target)
					if err != nil {

					}
					ans <- *ansFindNode
					alreadyContacted = append(alreadyContacted, contact)
				}
			}()
		}

		wg.Wait()
		var new_candidates []Contact

		for _ = range len(ans) {
			candidatesAns := <-ans
			for i := range len(candidatesAns.double) {
				new_candidate := candidatesAns.double[i]
				new_candidates = append(new_candidates, NewContact(new_candidate.id, new_candidate.address))
			}
		}

		candidates.Append(new_candidates)
		candidates.Sort()

		newClosestNode := candidates.GetContact(0)

		if newClosestNode == closestNode {
			noNewClosest = true
		}

		closestNode = newClosestNode
		candidates.PopShortList(k)

		probed = 0
		for i := range len(candidates.contacts) {
			if slices.Contains(alreadyContacted, candidates.GetContact(i)) == true {
				probed += 1
			}
		}
	}

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
	closest := kademlia.LookupContact(targetID).contacts

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
	for _, contact := range targetNodes.contacts {
		// go kademlia.Network.SendStoreRPC(&contact, keyHex, data)
		_ = contact
	}

	return keyHex
}
