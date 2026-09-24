package kademlia

import (
	"errors"
	"kademlia/internal/adapters"
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

const k_const = 4
const alpha = 3
const b = 1
const timeout = 1000 // ms

type Kademlia interface {
	Run() error
	Ping(address entities.Address) (time.Duration, error)
	GetBuckets() []*bucket
	GetStoredKeys() []string
}

type kademlia struct {
	RoutingTable *RoutingTable
	network      ports.Network
	dataStore    *DataStore
	Connection   ports.ListenConnection
	me           Contact
	mux          sync.RWMutex
}

// NewKademlia creates and initializes a new instance of the Kademlia node
func NewKademlia(me Contact, net ports.Network) *kademlia {
	return &kademlia{
		RoutingTable: NewRoutingTable(me),
		network:      net,
		dataStore:    NewDataStore(),
		me:           me,
	}
}

func (k *kademlia) Run() error {
	connection, err := k.network.Listen(k.me.Address)
	if err != nil {
		return err
	}
	k.mux.Lock()
	k.Connection = connection
	k.mux.Unlock()
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
			if err == adapters.ErrConnectionClosed {
				break
			} else {
				continue
			}
		}
		k.handleRequest(connection, payload, *address)
	}
	return nil
}

func (k *kademlia) Quit() error {
	k.mux.RLock()
	if k.Connection == nil {
		return errors.New("No connection")
	}
	err := k.Connection.Close()
	k.mux.RUnlock()
	if err != nil {
		return err
	}
	return nil
}

func (k *kademlia) UpdateRoutingTable(
	id KademliaID,
	address entities.Address,
) {
	bucketIndex := k.RoutingTable.getBucketIndex(&id)
	bucket := k.RoutingTable.buckets[bucketIndex]
	contact := bucket.AddContact(NewContact(&id, address))
	if contact != nil {
		oldestContact := bucket.list.Back().Value.(Contact)
		_, err := k.Ping(oldestContact.Address)
		if err != nil {
			slog.Info("Old contact not responding")
			bucket.RemoveContact(oldestContact)
			bucket.AddContact(*contact)
		} else {
			slog.Info("Contact not added")
		}
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

	k.UpdateRoutingTable(KademliaID(message.KademliaId), address)
	switch payload := message.Payload.(type) {
	case *generated.Message_Ping:
		k.handlePing(connection, payload.Ping, address)
	case *generated.Message_FindNode:
		k.handleFindNode(connection, payload.FindNode, address)
	}
}

func (k *kademlia) handleFindNode(
	connection ports.ListenConnection,
	findNode *generated.FindNode,
	address entities.Address,
) {
	slog.Info(
		"Receive find node message",
		"requestTarget",
		findNode.GetTargetId(),
		"requestRequester",
		findNode.GetRequesterId(),
		"requestRecipient",
		findNode.GetRecipientId(),
		"from",
		address,
	)

	var candidates_raw ContactCandidates
	candidates_raw.Append(k.RoutingTable.FindClosestContacts((*KademliaID)(findNode.GetTargetId()), k_const+1))
	candidates_raw.Sort()

	requester_ID := (*KademliaID)(findNode.GetRequesterId())
	var candidates ContactCandidates
	for i := range len(candidates_raw.contacts) {
		slog.Debug("IDs", "candidates_raw.contacts[i].ID", candidates_raw.contacts[i].ID, "requester_ID", requester_ID)
		if *candidates_raw.contacts[i].ID != *requester_ID {
			candidates.contacts = append(candidates.contacts, candidates_raw.contacts[i])
		}
	}

	if len(candidates.contacts) > k_const {
		candidates.PopShortList(k_const)
	}

	var findNodeResponse generated.FindNodeResponse

	for i := range len(candidates.contacts) {
		triple := generated.Triples{
			Address:    candidates.contacts[i].Address.IP,
			Port:       int32(candidates.contacts[i].Address.Port),
			Kademliaid: []byte(candidates.contacts[i].ID.String()),
		}
		findNodeResponse.Triples = append(findNodeResponse.Triples, &triple)
	}

	payload, err := proto.Marshal(&findNodeResponse)
	if err != nil {
		slog.Error("error", "err", err)
		return
	}

	if err := connection.SendTo(address, payload); err != nil {
		slog.Error("error", "err", err)
		return
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

func ParalelFindNode(req_id *KademliaID, kNet ports.Network, contact Contact, target *KademliaID, ans chan RPCResponse, remove chan Contact, wg *sync.WaitGroup) error {
	ansFindNode, err := SendFindNode(req_id, kNet, contact, target)
	if err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) {
			remove <- contact
		} else {
			wg.Done()
			return err
		}
	} else {
		slog.Debug("Finded Nodes", "ansFindNode", ansFindNode)
		ans <- *ansFindNode
	}
	wg.Done()
	return nil
}

// LookupContact does an iterative search of the closest k nodes to a target ID
func (k *kademlia) LookupContact(target *KademliaID) (*ContactCandidates, error) {
	// 1. Obtain the initial closest contacts from the local routing table
	var candidates ContactCandidates
	var noNewClosest bool
	var probed int
	// TODO: No RPC response reaction,

	var alreadyContacted []Contact

	candidates.Append(k.RoutingTable.FindClosestContacts(target, k_const))
	candidates.Sort()
	closestNode := candidates.GetContact(0)
	noNewClosest = false

	slog.Debug("Before Loop", "noNewClosest", noNewClosest, "probed", probed, "k_const", k_const)
	for (!noNewClosest) && (probed != k_const) {
		var wg sync.WaitGroup
		ans := make(chan RPCResponse, alpha)
		remove := make(chan Contact, alpha)
		slog.Debug("In Loop", "noNewClosest", noNewClosest, "probed", probed, "k_const", k_const)

		for nodeCounter := range min(alpha, candidates.Len()) {
			slog.Debug("Counter", "nodeCounter", nodeCounter, "candidates", candidates.Len())
			contact := candidates.GetContact(nodeCounter)
			if slices.Contains(alreadyContacted, contact) == false {
				k.mux.RLock()
				req_id := k.me.ID
				me_net := k.network
				k.mux.RUnlock()
				wg.Add(1)
				go ParalelFindNode(req_id, me_net, contact, target, ans, remove, &wg)
				alreadyContacted = append(alreadyContacted, contact)
			}
		}

		wg.Wait()
		for range len(remove) {
			to_remove := <-remove
			k.RoutingTable.RemoveContact(to_remove)
		}

		var new_candidates []Contact
		var new_candidates_id []KademliaID

		for range len(ans) {
			candidatesAns := <-ans
			for i := range len(candidatesAns.double) {
				new_candidate := candidatesAns.double[i]
				new_contact := NewContact(new_candidate.id, new_candidate.address)
				new_contact.CalcDistance(target)
				if !slices.Contains(new_candidates_id, *new_contact.ID) {
					new_candidates = append(new_candidates, new_contact)
					new_candidates_id = append(new_candidates_id, *new_contact.ID)
					slog.Debug("Candidates ID", "new_candidates_id", new_candidates_id)
				}
			}
		}

		var candidates_id []KademliaID

		for i := range len(candidates.contacts) {
			candidates_id = append(candidates_id, *candidates.contacts[i].ID)
		}

		slog.Debug("New candidates before removing", "new_candidates", new_candidates)
		var new_candidates_without_candidates []Contact

		for i := range len(new_candidates) {
			if !slices.Contains(candidates_id, *new_candidates[i].ID) {
				new_candidates_without_candidates = append(new_candidates_without_candidates, new_candidates[i])
			}
		}

		candidates.Append(new_candidates_without_candidates)
		slog.Debug("Candidates after find_node", "candidates", candidates)
		candidates.Sort()
		slog.Debug("Candidates after sort", "candidates", candidates)
		for i := range len(candidates.contacts) {
			slog.Debug("Distance", "distance", candidates.contacts[i].distance)
		}
		newClosestNode := candidates.GetContact(0)

		if newClosestNode == closestNode {
			noNewClosest = true
		}

		closestNode = newClosestNode
		if candidates.Len() > k_const {
			candidates.PopShortList(k_const)
		}

		probed = 0
		for i := range len(candidates.contacts) {
			if slices.Contains(alreadyContacted, candidates.GetContact(i)) == true {
				probed += 1
			}
		}
	}

	return &candidates, nil
}

func (k *kademlia) Ping(address entities.Address) (time.Duration, error) {
	connection, err := k.network.Dial(address)
	if err != nil {
		return 0, err
	}
	defer connection.Close()

	requestUuid := uuid.New().String()

	pingMessage := generated.Message{
		KademliaId: k.me.ID[:],
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

func (k *kademlia) GetBuckets() []*bucket {
	return k.RoutingTable.getBuckets()
}

func (k *kademlia) GetStoredKeys() []string {
	return k.dataStore.Keys()
}
