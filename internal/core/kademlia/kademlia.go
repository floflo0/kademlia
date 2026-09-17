package kademlia

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"kademlia/internal/core/entities"
	"kademlia/internal/core/ports"
	"kademlia/proto/generated"
	"log/slog"
	"slices"
	"sync"
	"syscall"
	"time"
	"uuid"

	"google.golang.org/protobuf/proto"
)

const k = 4
const alpha = 3
const b = 1
const timeout = 1000 // ms

type Kademlia interface {
	Run() error
	Ping(address entities.Address) (time.Duration, error)
}

type kademlia struct {
	RoutingTable *RoutingTable
	network      ports.Network
	DataStore    map[string][]byte
	me           Contact
	mux          sync.RWMutex
}

// NewKademlia creates and initializes a new instance of the Kademlia node
func NewKademlia(me Contact, net ports.Network) *kademlia {
	slog.Debug("", "me", me)
	return &kademlia{
		RoutingTable: NewRoutingTable(me),
		network:      net,
		DataStore:    make(map[string][]byte),
		me:           me,
	}
}

func (k *kademlia) Run() error {
	connection, err := k.network.Listen(k.me.Address)
	if err != nil {
		return err
	}
	defer connection.Close()
	ip, err := connection.GetIP()
	if err != nil {
		return err
	}
	slog.Info("Server started", "ip", ip, "port", k.me.Address.Port)

	for {
		payload, address, err := connection.Receive()
		if err != nil {
			slog.Error("todo: message", "err", err)
			continue
		}
		k.handleRequest(connection, payload, *address)
	}
}

func (k *kademlia) handleRequest(
	connection ports.ListenConnection,
	payload []byte,
	address entities.Address,
) {
	message := generated.Message{}
	if err := proto.Unmarshal(payload, &message); err != nil {
		slog.Error("Error while unmarshaling request payload", "err", err)
		return
	}

	switch payload := message.Payload.(type) {
	case *generated.Message_Ping:
		k.handlePing(connection, payload.Ping, address)
	}
}

func (k *kademlia) handlePing(
	connection ports.ListenConnection,
	ping *generated.Ping,
	address entities.Address,
) {
	slog.Info(
		"Receive ping message",
		"requestUuid",
		ping.RequestUuid,
		"from",
		address,
	)
	pongMessage := generated.Pong{
		RequestUuid: ping.RequestUuid,
	}
	payload, err := proto.Marshal(&pongMessage)
	if err != nil {
		slog.Error("error", "err", err)
		return
	}

	if err := connection.SendTo(address, payload); err != nil {
		slog.Error("error", "err", err)
		return
	}
}

// LookupContact does an iterative search of the closest k nodes to a target ID
func (kademlia *kademlia) LookupContact(target *KademliaID) (*ContactCandidates, error) {
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

		for nodeCounter := range alpha {
			wg.Add(1)

			go func() error {
				defer wg.Done()
				contact := candidates.GetContact(nodeCounter)
				if slices.Contains(alreadyContacted, contact) == false {
					ansFindNode, err := SendFindNode(kademlia.network, contact.Address, target)
					if err != nil {
						if errors.Is(err, syscall.ECONNREFUSED) {
							kademlia.RoutingTable.RemoveContact(contact)
						} else {
							return err
						}
					}
					ans <- *ansFindNode
					alreadyContacted = append(alreadyContacted, contact)
				}
				return nil
			}()
		}

		wg.Wait()
		var new_candidates []Contact

		for range len(ans) {
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

	return &candidates, nil
}

// LookupData searches for the value belonging to a key (hash)
// If it finds the value locally, it returns (data, nil, true)
// If it doesn't, it returns (nil, kClosestContacts, false)
func (kademlia *kademlia) LookupData(hash string) ([]byte, []Contact, bool) {
	kademlia.mux.RLock()
	val, exists := kademlia.DataStore[hash]
	kademlia.mux.RUnlock()

	// If the value is stored locally, return it immediately
	if exists {
		return val, nil, true
	}

	targetID := NewKademliaID(hash)
	closestCandidates, err := kademlia.LookupContact(targetID)
	if err != nil {

	}

	// TODO: Send RPCs FIND_VALUE to the closest nodes until
	// obtaining the value or running out of contacts.

	return nil, closestCandidates.contacts, false
}

// Store calculates the SHA-256 key of the data, saves it locally,
// searches for the closest k nodes, and sends a STORE RPC to each one.
func (kademlia *kademlia) Store(data []byte) string {
	// 1. Calculate K = SHA-256(data) (32 bytes / 64 hex char to match IDLength=32)
	hashBytes := sha256.Sum256(data)
	keyHex := hex.EncodeToString(hashBytes[:])
	keyID := NewKademliaID(keyHex)

	// 2. Save a copy in the local DataStore safely
	kademlia.mux.Lock()
	kademlia.DataStore[keyHex] = data
	kademlia.mux.Unlock()

	// 3. Search for the closest k nodes to the key
	targetNodes, err := kademlia.LookupContact(keyID)
	if err != nil {

	}

	// 4. Send a STORE RPC to each of the closest k nodes
	for _, contact := range targetNodes.contacts {
		// go kademlia.Network.SendStoreRPC(&contact, keyHex, data)
		_ = contact
	}

	return keyHex
}

func (k *kademlia) Ping(address entities.Address) (time.Duration, error) {
	connection, err := k.network.Dial(address)
	if err != nil {
		return 0, err
	}
	defer connection.Close()

	requestUuid := uuid.New().String()

	pingMessage := generated.Message{
		Payload: &generated.Message_Ping{
			Ping: &generated.Ping{
				RequestUuid: requestUuid,
			},
		},
	}
	payload, err := proto.Marshal(&pingMessage)
	if err != nil {
		return 0, err
	}

	startTime := time.Now()
	if err = connection.Send(payload); err != nil {
		return 0, err
	}

	payload, err = connection.Receive(timeout)
	if err != nil {
		return 0, err
	}
	elapsedTime := time.Since(startTime)

	pongMessage := generated.Pong{}
	if err = proto.Unmarshal(payload, &pongMessage); err != nil {
		return 0, err
	}

	if requestUuid != pongMessage.RequestUuid {
		return 0, errors.New("invalid request uuid")
	}

	return elapsedTime, nil
}
