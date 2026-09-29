package kademlia

import (
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

const k_const = 4
const alpha = 3
const b = 1
const timeout = 1000 // ms

var ErrDataLen = errors.New("Lenght of data does not correspond to Data lenght sent")

type Kademlia interface {
	Run(firstContact *entities.Address) error
	Ping(address entities.Address) (time.Duration, error)
	GetBuckets() []*Bucket
	GetStoredKeys() []string
	GetValue(key string) (*string, *string, error)
	Quit() error
}

type kademlia struct {
	RoutingTable *RoutingTable
	network      ports.Network
	dataStore    *DataStore
	connection   ports.ListenConnection
	firstContact *entities.Address
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

func (k *kademlia) join(knownContact *entities.Address) error {
	// Already has a NodeId cause we made it mandatory to create a NewKademlia
	id := NewKademliaIDFomAddress(*knownContact)
	slog.Info("ID finded with IP|port combination", "id", id, "k.me.ID", k.me.ID)
	slog.Debug("New Kademlia ID generated from Address", "id", id)
	k.RoutingTable.AddContact(NewContact(
		id,
		*knownContact,
	))
	_, err := k.LookupContact(k.me.ID)
	return err
}

func (k *kademlia) Run(firstContact *entities.Address) error {
	connection, err := k.network.Listen(k.me.Address)
	if err != nil {
		return err
	}
	defer connection.Close()
	k.mux.Lock()
	k.connection = connection
	k.mux.Unlock()
	ip, err := connection.GetIP()
	if err != nil {
		return err
	}
	slog.Info("Server started", "ip", ip, "port", k.me.Address.Port)

	if firstContact != nil {
		slog.Info("Joining the Kademlia network")
		err := k.join(firstContact)
		if err != nil {
			return err
		}
	}

	for {
		payload, address, err := connection.Receive()
		if err != nil {
			if errors.Is(err, ports.ErrClosedNetworkConnection) {
				break
			}
			slog.Error("failed to receive packet", "err", err)
			continue
		}
		k.handleRequest(connection, payload, *address)
	}
	return nil
}

func (k *kademlia) Quit() error {
	k.mux.Lock()
	defer k.mux.Unlock()
	if k.connection == nil {
		return errors.New("can't quit Kademlia: connection is nil")
	}
	err := k.connection.Close()
	if err != nil {
		return err
	}
	k.connection = nil
	return nil
}

func (k *kademlia) UpdateRoutingTable(
	id KademliaID,
	address entities.Address,
) {
	k.RoutingTable.AddContact(NewContact(&id, address))
}

func (k *kademlia) handleRequest(
	connection ports.ListenConnection,
	payload []byte,
	address entities.Address,
) {
	message := generated.Message{}
	if err := proto.Unmarshal(payload, &message); err != nil {
		slog.Error("Error while unmarshaling request payload", "err", err)
		slog.Info("Error")
		return
	}

	k.UpdateRoutingTable(KademliaID(message.KademliaId), address)
	switch payload := message.Payload.(type) {
	case *generated.Message_Ping:
		k.handlePing(connection, payload.Ping, address)
	case *generated.Message_FindNode:
		k.handleFindNode(connection, payload.FindNode, address)
	case *generated.Message_FindValue:
		k.handleFindValue(connection, payload.FindValue, address)
	case *generated.Message_Store:
		k.handleStore(payload.Store, address)
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

	var findNodeTriples generated.FindNodeResponse

	for i := range len(candidates.contacts) {
		triple := generated.Triples{
			Address:    candidates.contacts[i].Address.IP,
			Port:       int32(candidates.contacts[i].Address.Port),
			Kademliaid: []byte(candidates.contacts[i].ID.String()),
		}
		findNodeTriples.Triples = append(findNodeTriples.Triples, &triple)
	}

	findNodeResponse := generated.Message{
		KademliaId: k.me.ID[:],
		Payload: &generated.Message_FindNodeReponse{
			FindNodeReponse: &findNodeTriples,
		},
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

func (k *kademlia) handleFindValue(
	connection ports.ListenConnection,
	findValue *generated.FindValue,
	address entities.Address,
) {
	slog.Info(
		"Receive find node message",
		"requestTarget",
		findValue.GetTargetId(),
		"requestRequester",
		findValue.GetRequesterId(),
		"requestRecipient",
		findValue.GetRecipientId(),
		"from",
		address,
	)
	target := (*KademliaID)(findValue.GetTargetId())
	if slices.Contains(k.dataStore.Keys(), target.String()) {
		value := k.dataStore.data[target.String()]
		findValueResponse := generated.Message{
			KademliaId: k.me.ID[:],
			Payload: &generated.Message_FindValueResponse{
				FindValueResponse: &generated.FindValueResponse{
					Triples: nil,
					Value:   value,
				},
			},
		}

		payload, err := proto.Marshal(&findValueResponse)
		if err != nil {
			slog.Error("error", "err", err)
			return
		}

		if err := connection.SendTo(address, payload); err != nil {
			slog.Error("error", "err", err)
			return
		}
		return
	}

	var candidates_raw ContactCandidates
	candidates_raw.Append(k.RoutingTable.FindClosestContacts(target, k_const+1))
	candidates_raw.Sort()

	requester_ID := (*KademliaID)(findValue.GetRequesterId())
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

	var findValueTriple generated.FindValueResponse

	for i := range len(candidates.contacts) {
		triple := generated.Triples{
			Address:    candidates.contacts[i].Address.IP,
			Port:       int32(candidates.contacts[i].Address.Port),
			Kademliaid: []byte(candidates.contacts[i].ID.String()),
		}
		findValueTriple.Triples = append(findValueTriple.Triples, &triple)
	}

	findValueTriple.Value = ""

	findValueResponse := generated.Message{
		KademliaId: k.me.ID[:],
		Payload: &generated.Message_FindValueResponse{
			FindValueResponse: &findValueTriple,
		},
	}

	payload, err := proto.Marshal(&findValueResponse)
	if err != nil {
		slog.Error("error", "err", err)
		return
	}

	if err := connection.SendTo(address, payload); err != nil {
		slog.Error("error", "err", err)
		return
	}
}

func (k *kademlia) handleStore(
	store *generated.Store,
	address entities.Address,
) error {
	slog.Info(
		"Receive Store message",
		"key",
		store.GetKey(),
		"requestRequester",
		store.GetRequesterId(),
		"requestRecipient",
		store.GetRecipientId(),
		"from",
		address,
	)
	key := store.GetKey()
	data := store.GetData()
	dataLen := store.GetDataLen()

	if len(data) != int(dataLen) {
		return ErrDataLen
	}

	k.dataStore.Put(string(key), string(data))

	return nil
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

func ParalelFindNode(req_id *KademliaID, kNet ports.Network, contact Contact, target *KademliaID, ans chan RPCResponseNode, remove chan Contact, wg *sync.WaitGroup) error {
	defer wg.Done()
	ansFindNode, err := SendFindNode(req_id, kNet, contact, target)
	if err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) {
			remove <- contact
		} else {
			return err
		}
	} else {
		slog.Debug("Finded Nodes", "ansFindNode", ansFindNode)
		ans <- *ansFindNode
	}
	return nil
}

func ParalelFindValue(req_id *KademliaID, kNet ports.Network, contact Contact, target *KademliaID, ans chan RPCResponseValue, remove chan Contact, wg *sync.WaitGroup) error {
	defer wg.Done()
	ansFindValue, err := SendFindValue(req_id, kNet, contact, target)
	if err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) {
			remove <- contact
		} else {
			return err
		}
	} else {
		slog.Debug("Finded Values/Nodes", "ansFindValue", ansFindValue)
		ans <- *ansFindValue
	}
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
		ans := make(chan RPCResponseNode, alpha)
		remove := make(chan Contact, alpha)
		slog.Debug("In Loop", "noNewClosest", noNewClosest, "probed", probed, "k_const", k_const)

		for nodeCounter := range min(alpha, candidates.Len()) {
			slog.Debug("Counter", "nodeCounter", nodeCounter, "candidates", candidates.Len())
			contact := candidates.GetContact(nodeCounter)
			if slices.Contains(alreadyContacted, contact) == false {
				alreadyContacted = append(alreadyContacted, contact)
				k.mux.RLock()
				req_id := k.me.ID
				me_net := k.network
				k.mux.RUnlock()
				wg.Add(1)
				go ParalelFindNode(req_id, me_net, contact, target, ans, remove, &wg)
			}
		}

		wg.Wait()
		var removed []Contact
		for range len(remove) {
			to_remove := <-remove
			k.RoutingTable.RemoveContact(to_remove)
			removed = append(removed, to_remove)
		}

		for i := range len(removed) {
			for j := range len(candidates.contacts) {
				if removed[i] == candidates.contacts[j] {
					candidates.contacts = append(candidates.contacts[:j], candidates.contacts[j+1:]...)
				}
			}
		}

		// For race condition, as we have removed all the node that didn't answered
		// we only have the nodes that responded so we have to update the routing table
		for i := range len(candidates.contacts) {
			k.UpdateRoutingTable(*candidates.contacts[i].ID, candidates.contacts[i].Address)
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
	slog.Info("Stopping the Lookup Loop", "!noNewClosest", !noNewClosest, "probed != k_const", probed != k_const)
	slog.Info("State of bucket", "k.RoutingTable.FindClosestContacts(k.me.ID, 10)", k.RoutingTable.FindClosestContacts(k.me.ID, 10))

	return &candidates, nil
}

// LookupContact does an iterative search of the closest k nodes to a target ID
func (k *kademlia) LookupValue(target *KademliaID) (*string, *Contact, *ContactCandidates, error) {
	// 1. Obtain the initial closest contacts from the local routing table
	var candidates ContactCandidates
	var noNewClosest bool
	var probed int
	var answer *string
	var candidatesWithAnswer []KademliaID
	valueFound := false

	var alreadyContacted []Contact

	if slices.Contains(k.dataStore.Keys(), target.String()) {
		value := k.dataStore.data[target.String()]
		return &value, &k.me, nil, nil
	}

	candidates.Append(k.RoutingTable.FindClosestContacts(target, k_const))
	candidates.Sort()
	closestNode := candidates.GetContact(0)
	noNewClosest = false

	slog.Debug("Before Loop", "noNewClosest", noNewClosest, "probed", probed, "k_const", k_const)
	for (!noNewClosest) && (probed != k_const) && !valueFound {
		var wg sync.WaitGroup
		ans := make(chan RPCResponseValue, alpha)
		remove := make(chan Contact, alpha)
		slog.Debug("In Loop", "noNewClosest", noNewClosest, "probed", probed, "k_const", k_const)

		for nodeCounter := range min(alpha, candidates.Len()) {
			slog.Debug("Counter", "nodeCounter", nodeCounter, "candidates", candidates.Len())
			contact := candidates.GetContact(nodeCounter)
			if slices.Contains(alreadyContacted, contact) == false {
				alreadyContacted = append(alreadyContacted, contact)
				k.mux.RLock()
				req_id := k.me.ID
				me_net := k.network
				k.mux.RUnlock()
				wg.Add(1)
				go ParalelFindValue(req_id, me_net, contact, target, ans, remove, &wg)
			}
		}

		wg.Wait()
		var removed []Contact
		for range len(remove) {
			to_remove := <-remove
			k.RoutingTable.RemoveContact(to_remove)
			removed = append(removed, to_remove)
		}

		for i := range len(removed) {
			for j := range len(candidates.contacts) {
				if removed[i] == candidates.contacts[j] {
					candidates.contacts = append(candidates.contacts[:j], candidates.contacts[j+1:]...)
				}
			}
		}

		// For race condition, as we have removed all the node that didn't answered
		// we only have the nodes that responded so we have to update the routing table
		for i := range len(candidates.contacts) {
			k.UpdateRoutingTable(*candidates.contacts[i].ID, candidates.contacts[i].Address)
		}

		var new_candidates []Contact
		var new_candidates_id []KademliaID

		for range len(ans) {
			candidatesAns := <-ans
			if candidatesAns.value != nil && *candidatesAns.value != "" {
				valueFound = true
				answer = candidatesAns.value
				candidatesWithAnswer = append(candidatesWithAnswer, *candidatesAns.sender)
			} else {
				double := *candidatesAns.double
				for i := range len(double) {
					new_candidate := double[i]
					new_contact := NewContact(new_candidate.id, new_candidate.address)
					new_contact.CalcDistance(target)
					if !slices.Contains(new_candidates_id, *new_contact.ID) {
						new_candidates = append(new_candidates, new_contact)
						new_candidates_id = append(new_candidates_id, *new_contact.ID)
						slog.Debug("Candidates ID", "new_candidates_id", new_candidates_id)
					}
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
		slog.Info("List of candidates in order", "candidates.contacts", candidates.contacts)

		probed = 0
		for i := range len(candidates.contacts) {
			if slices.Contains(alreadyContacted, candidates.GetContact(i)) == true {
				probed += 1
			}
		}
	}
	slog.Info("Stopping the Lookup Loop", "!noNewClosest", !noNewClosest, "probed != k_const", probed != k_const, "valueFound", valueFound)
	slog.Info("State of bucket", "k.RoutingTable.FindClosestContacts(k.me.ID, 10)", k.RoutingTable.FindClosestContacts(k.me.ID, 10))
	var contactWithAnswer Contact
	if answer != nil {
		for i := range len(candidates.contacts) {
			if !slices.Contains(candidatesWithAnswer, *candidates.contacts[i].ID) {
				SendStore(k.me.ID, k.network, candidates.contacts[i], target.String(), *answer)
			} else {
				contactWithAnswer = candidates.contacts[i]
				break
			}
		}
		return answer, &contactWithAnswer, nil, nil
	}

	return nil, nil, &candidates, nil
}

func (k *kademlia) Store(key string, data string) error {
	candidates, err := k.LookupContact(NewKademliaID(key))
	if err != nil {
		return err
	}

	for i := range len(candidates.contacts) {
		SendStore(k.me.ID, k.network, candidates.contacts[i], key, data)
	}

	return nil
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

func (k *kademlia) GetBuckets() []*Bucket {
	return k.RoutingTable.getBuckets()
}

func (k *kademlia) GetStoredKeys() []string {
	return k.dataStore.Keys()
}

func (k *kademlia) GetValue(key string) (*string, *string, error) {
	target := NewKademliaID(key)
	value, contact, _, err := k.LookupValue(target)
	if err != nil {
		return nil, nil, err

	}
	if value != nil {
		contactID := contact.ID.String()
		return value, &contactID, nil
	}
	return nil, nil, nil
}
