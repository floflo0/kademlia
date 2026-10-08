package kademlia

import (
	"crypto/ed25519"
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

const K = 15
const alpha = 3
const b = 1
const timeout = 1000 // ms

var ErrDataLen = errors.New("lenght of data does not correspond to Data lenght sent")
var ErrNoContact = errors.New("no contact")

type Kademlia interface {
	Run(firstContact *entities.Address) error
	Ping(address entities.Address) (time.Duration, error)
	Put(data string) (*entities.KademliaID, error)
	GetBuckets() []*Bucket
	GetStoredKeys() []entities.KademliaID
	GetValue(key string) (*string, *string, error)
	Quit() error

	// Package Management (Part 2)
	PublishPackage(domain, packageName, version, blob string, privKey ed25519.PrivateKey, dns DNSVerifier) (*VersionRecord, error)
	InstallPackage(domain, packageName, version string) (string, string, error)
}

type kademlia struct {
	RoutingTable *RoutingTable
	network      ports.Network
	dataStore    ports.DataStore
	connection   ports.ListenConnection
	me           Contact
	mux          sync.RWMutex
}

// NewKademlia creates and initializes a new instance of the Kademlia node
func NewKademlia(address entities.Address, network ports.Network) *kademlia {
	me := NewContactFromAddress(address)
	return &kademlia{
		RoutingTable: NewRoutingTable(me),
		network:      network,
		dataStore:    adapters.NewInMemoryDataStore(),
		me:           me,
	}
}

func (k *kademlia) join(knownContact *entities.Address) error {
	// Already has a NodeId cause we made it mandatory to create a NewKademlia
	id := entities.NewKademliaIDFromAddress(*knownContact)
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
	slog.Info("Server started", "ip", k.me.Address.IP, "port", k.me.Address.Port)

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
		go k.handleRequest(connection, payload, *address)
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

func (k *kademlia) updateRoutingTable(
	id entities.KademliaID,
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
		return
	}

	requesterID := (entities.KademliaID)(message.Contact.KademliaId)
	k.updateRoutingTable(
		requesterID,
		entities.Address{
			IP:   message.Contact.Ip,
			Port: int(message.Contact.Port),
		},
	)
	switch payload := message.Payload.(type) {
	case *generated.Message_Ping:
		k.handlePing(connection, payload.Ping, address)
	case *generated.Message_FindNode:
		k.handleFindNode(connection, payload.FindNode, address, requesterID)
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
	requesterID entities.KademliaID,
) error {
	key := (entities.KademliaID)(findNode.GetKey())
	slog.Info(
		"Receive FIND_NODE RPC",
		"key",
		key.String(),
		"from",
		address,
	)

	var candidatesRaw ContactCandidates
	candidatesRaw.Append(k.RoutingTable.FindClosestContacts(&key, K+1))

	var candidates ContactCandidates
	for i := range len(candidatesRaw.contacts) {
		if *candidatesRaw.contacts[i].ID != requesterID {
			candidates.contacts = append(candidates.contacts, candidatesRaw.contacts[i])
		}
	}

	if len(candidates.contacts) > K {
		candidates.PopShortList(K)
	}

	var findNodeTriples generated.FindNodeResponse

	for i := range len(candidates.contacts) {
		slog.Debug("Find node response", "contactID", candidatesRaw.contacts[i].ID)
		triple := generated.Triples{
			Address:    candidates.contacts[i].Address.IP,
			Port:       int32(candidates.contacts[i].Address.Port),
			Kademliaid: []byte(candidates.contacts[i].ID.String()),
		}
		findNodeTriples.Triples = append(findNodeTriples.Triples, &triple)
	}

	findNodeResponse := generated.Message{
		Contact: &generated.Contact{
			KademliaId: k.me.ID[:],
			Ip:         k.me.Address.IP,
			Port:       int32(k.me.Address.Port),
		},
		Payload: &generated.Message_FindNodeReponse{
			FindNodeReponse: &findNodeTriples,
		},
	}

	payload, err := proto.Marshal(&findNodeResponse)
	if err != nil {
		slog.Error("error", "err", err)
		return err
	}

	if err := connection.SendTo(address, payload); err != nil {
		slog.Error("error", "err", err)
		return err
	}
	return nil
}

func (k *kademlia) handleFindValue(
	connection ports.ListenConnection,
	findValue *generated.FindValue,
	address entities.Address,
) error {
	target := (*entities.KademliaID)(findValue.GetTargetId())
	slog.Info(
		"Receive find value message",
		"requestTarget",
		target.String(),
		"requestRequester",
		findValue.GetRequesterId(),
		"requestRecipient",
		findValue.GetRecipientId(),
		"from",
		address,
	)
	if value, err := k.dataStore.Get(*target); err == nil {
		findValueResponse := generated.Message{
			Contact: &generated.Contact{
				KademliaId: k.me.ID[:],
				Ip:         k.me.Address.IP,
				Port:       int32(k.me.Address.Port),
			},
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
			return err
		}

		if err := connection.SendTo(address, payload); err != nil {
			slog.Error("error", "err", err)
			return err
		}
		return nil
	}

	var candidatesRaw ContactCandidates
	candidatesRaw.Append(k.RoutingTable.FindClosestContacts(target, K+1))
	candidatesRaw.Sort()

	requesterID := (*entities.KademliaID)(findValue.GetRequesterId())
	var candidates ContactCandidates
	for i := range len(candidatesRaw.contacts) {
		slog.Debug("IDs", "candidates_raw.contacts[i].ID", candidatesRaw.contacts[i].ID, "requester_ID", requesterID)
		if *candidatesRaw.contacts[i].ID != *requesterID {
			candidates.contacts = append(candidates.contacts, candidatesRaw.contacts[i])
		}
	}

	if len(candidates.contacts) > K {
		candidates.PopShortList(K)
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
		Contact: &generated.Contact{
			KademliaId: k.me.ID[:],
			Ip:         k.me.Address.IP,
			Port:       int32(k.me.Address.Port),
		},
		Payload: &generated.Message_FindValueResponse{
			FindValueResponse: &findValueTriple,
		},
	}

	payload, err := proto.Marshal(&findValueResponse)
	if err != nil {
		slog.Error("error", "err", err)
		return err
	}

	if err := connection.SendTo(address, payload); err != nil {
		slog.Error("error", "err", err)
		return err
	}

	return nil
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

	id := (*entities.KademliaID)(key)
	k.dataStore.Put(*id, string(data))

	return nil
}

func (k *kademlia) handlePing(
	connection ports.ListenConnection,
	ping *generated.Ping,
	address entities.Address,
) error {
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
		return err
	}

	if err := connection.SendTo(address, payload); err != nil {
		slog.Error("error", "err", err)
		return err
	}
	return nil
}

func (k *kademlia) ParalelFindNode(
	reqID *entities.KademliaID,
	kNet ports.Network,
	contact Contact,
	target *entities.KademliaID,
	ans chan RPCResponseNode,
	remove chan Contact,
	wg *sync.WaitGroup,
) error {
	defer wg.Done()
	ansFindNode, err := k.SendFindNode(contact, target)
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

func (k *kademlia) ParalelFindValue(
	reqID *entities.KademliaID,
	kNet ports.Network,
	contact Contact,
	target *entities.KademliaID,
	ans chan RPCResponseValue,
	remove chan Contact,
	wg *sync.WaitGroup,
) error {
	defer wg.Done()
	ansFindValue, err := k.SendFindValue(contact, target)
	if err != nil {
		slog.Debug("errror")
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
// return nil if the node is alone in the network
func (k *kademlia) LookupContact(
	target *entities.KademliaID,
) (*ContactCandidates, error) {
	var candidates ContactCandidates
	candidates.Append(k.RoutingTable.FindClosestContacts(target, K))
	if candidates.Len() == 0 {
		return nil, nil
	}

	var alreadyContacted []Contact
	var probed int
	closestNode := candidates.GetContact(0)
	noNewClosest := false

	for {
		ans := make(chan RPCResponseNode, alpha)
		remove := make(chan Contact, alpha)
		slog.Debug("In Loop", "noNewClosest", noNewClosest, "probed", probed)

		var toContact []Contact
		for i := 0; i < candidates.Len() && len(toContact) < alpha; i++ {
			c := candidates.GetContact(i)
			if !slices.Contains(alreadyContacted, c) {
				toContact = append(toContact, c)
			}
		}
		if len(toContact) == 0 {
			break
		}

		var waitGroup sync.WaitGroup
		for i := range toContact {
			contact := toContact[i]
			alreadyContacted = append(alreadyContacted, contact)
			waitGroup.Add(1)
			go k.ParalelFindNode(
				k.me.ID,
				k.network,
				contact,
				target,
				ans,
				remove,
				&waitGroup,
			)
		}
		waitGroup.Wait()

		var removed []Contact
		for range len(remove) {
			toRemove := <-remove
			k.RoutingTable.RemoveContact(toRemove)
			removed = append(removed, toRemove)
		}

		max := len(candidates.contacts)
		for i := range len(removed) {
			for j := 0; j < max; {
				if removed[i] == candidates.contacts[j] {
					slog.Info("Removing a contact from the candidate list")
					candidates.contacts = append(candidates.contacts[:j], candidates.contacts[j+1:]...)
					max = max - 1
					continue
				}
				j++
			}
		}

		// For race condition, as we have removed all the node that didn't answered
		// we only have the nodes that responded so we have to update the routing table
		for i := range len(candidates.contacts) {
			k.updateRoutingTable(*candidates.contacts[i].ID, candidates.contacts[i].Address)
		}

		var newCandidates []Contact
		var newCandidatesID []entities.KademliaID

		for range len(ans) {
			candidatesAns := <-ans
			slog.Info("Number of answer received", "len(candidatesAns.double)", len(candidatesAns.double))
			for i := range len(candidatesAns.double) {
				newCandidate := candidatesAns.double[i]
				newContact := NewContact(newCandidate.id, newCandidate.address)
				newContact.CalcDistance(target)
				if !slices.Contains(newCandidatesID, *newContact.ID) {
					newCandidates = append(newCandidates, newContact)
					newCandidatesID = append(newCandidatesID, *newContact.ID)
					slog.Debug("Candidates ID", "new_candidates_id", newCandidatesID)
				}
			}
		}

		var candidatesID []entities.KademliaID

		for i := range len(candidates.contacts) {
			candidatesID = append(candidatesID, *candidates.contacts[i].ID)
		}

		slog.Debug("New candidates before removing", "new_candidates", newCandidates)
		var newCandidatesWithoutPreviousCandidates []Contact

		for i := range len(newCandidates) {
			if !slices.Contains(candidatesID, *newCandidates[i].ID) {
				newCandidatesWithoutPreviousCandidates = append(newCandidatesWithoutPreviousCandidates, newCandidates[i])
			}
		}

		candidates.Append(newCandidatesWithoutPreviousCandidates)
		slog.Debug("Candidates after find_node", "candidates", candidates)
		if candidates.Len() == 0 {
			return nil, nil
		}
		candidates.Sort()
		slog.Debug("Candidates after sort", "candidates", candidates)
		for i := range len(candidates.contacts) {
			slog.Debug("Distance", "distance", candidates.contacts[i].distance)
		}
		newClosestNode := candidates.GetContact(0)

		if closestNode.ID != nil && newClosestNode.ID != nil && newClosestNode.ID.Equals(closestNode.ID) {
			noNewClosest = true
		}

		closestNode = newClosestNode
		if candidates.Len() > K {
			candidates.PopShortList(K)
		}

		probed = 0
		for i := range len(candidates.contacts) {
			if slices.Contains(alreadyContacted, candidates.contacts[i]) {
				probed += 1
			}
		}

		if noNewClosest && (probed >= K || probed >= candidates.Len()) {
			break
		}
	}
	slog.Info("Stopping the Lookup Loop", "!noNewClosest", !noNewClosest, "probed != k_const", probed != K)
	slog.Info("State of bucket", "k.RoutingTable.FindClosestContacts(k.me.ID, 10)", k.RoutingTable.FindClosestContacts(k.me.ID, 10))

	return &candidates, nil
}

// LookupContact does an iterative search of the closest k nodes to a target ID
func (k *kademlia) LookupValue(
	target *entities.KademliaID,
) (*string, *Contact, *ContactCandidates, error) {
	value, err := k.dataStore.Get(*target)
	if value != "" && err == nil {
		return &value, &k.me, nil, nil
	}

	var candidates ContactCandidates
	var noNewClosest bool
	var probed int
	var answer *string
	var candidatesWithAnswer []entities.KademliaID
	valueFound := false

	var alreadyContacted []Contact

	candidates.Append(k.RoutingTable.FindClosestContacts(target, K))
	candidates.Sort()
	if len(candidates.contacts) == 0 {
		return nil, nil, nil, ErrNoContact
	}
	closestNode := candidates.GetContact(0)
	noNewClosest = false

	slog.Debug("Before Loop", "noNewClosest", noNewClosest, "probed", probed, "k_const", K)
	for !valueFound {
		var wg sync.WaitGroup
		ans := make(chan RPCResponseValue, alpha)
		remove := make(chan Contact, alpha)
		slog.Debug("In Loop", "noNewClosest", noNewClosest, "probed", probed, "k_const", K)

		var toContact []Contact
		for i := 0; i < candidates.Len() && len(toContact) < alpha; i++ {
			c := candidates.GetContact(i)
			if !slices.Contains(alreadyContacted, c) {
				toContact = append(toContact, c)
			}
		}
		if len(toContact) == 0 {
			break
		}

		for i := range toContact {
			contact := toContact[i]
			alreadyContacted = append(alreadyContacted, contact)
			k.mux.RLock()
			reqID := k.me.ID
			meNet := k.network
			k.mux.RUnlock()
			wg.Add(1)
			go k.ParalelFindValue(reqID, meNet, contact, target, ans, remove, &wg)
		}

		wg.Wait()
		var removed []Contact
		for range len(remove) {
			toRemove := <-remove
			k.RoutingTable.RemoveContact(toRemove)
			removed = append(removed, toRemove)
		}

		for i := range len(removed) {
			for j := 0; j < len(candidates.contacts); {
				if removed[i] == candidates.contacts[j] {
					candidates.contacts = append(candidates.contacts[:j], candidates.contacts[j+1:]...)
					continue
				}
				j++
			}
		}

		// For race condition, as we have removed all the node that didn't answered
		// we only have the nodes that responded so we have to update the routing table
		for i := range len(candidates.contacts) {
			k.updateRoutingTable(*candidates.contacts[i].ID, candidates.contacts[i].Address)
		}

		var newCandidates []Contact
		var newCandidatesID []entities.KademliaID

		for range len(ans) {
			candidatesAns := <-ans
			if candidatesAns.value != nil && *candidatesAns.value != "" {
				valueFound = true
				answer = candidatesAns.value
				if candidatesAns.sender != nil {
					candidatesWithAnswer = append(candidatesWithAnswer, *candidatesAns.sender)
				}
			} else if candidatesAns.double != nil {
				double := *candidatesAns.double
				for i := range len(double) {
					newCandidate := double[i]
					newContact := NewContact(newCandidate.id, newCandidate.address)
					newContact.CalcDistance(target)
					if !slices.Contains(newCandidatesID, *newContact.ID) {
						newCandidates = append(newCandidates, newContact)
						newCandidatesID = append(newCandidatesID, *newContact.ID)
						slog.Debug("Candidates ID", "new_candidates_id", newCandidatesID)
					}
				}
			}
		}

		if valueFound {
			break
		}

		var candidatesID []entities.KademliaID

		for i := range len(candidates.contacts) {
			candidatesID = append(candidatesID, *candidates.contacts[i].ID)
		}

		slog.Debug("New candidates before removing", "new_candidates", newCandidates)
		var newCandidatesWithoutCandidates []Contact

		for i := range len(newCandidates) {
			if !slices.Contains(candidatesID, *newCandidates[i].ID) {
				newCandidatesWithoutCandidates = append(newCandidatesWithoutCandidates, newCandidates[i])
			}
		}

		candidates.Append(newCandidatesWithoutCandidates)
		slog.Debug("Candidates after find_node", "candidates", candidates)
		candidates.Sort()
		slog.Debug("Candidates after sort", "candidates", candidates)
		for i := range len(candidates.contacts) {
			slog.Debug("Distance", "distance", candidates.contacts[i].distance)
		}
		if candidates.Len() == 0 {
			return nil, nil, nil, nil
		}
		newClosestNode := candidates.GetContact(0)

		if closestNode.ID != nil && newClosestNode.ID != nil && newClosestNode.ID.Equals(closestNode.ID) {
			noNewClosest = true
		}

		closestNode = newClosestNode
		if candidates.Len() > K {
			candidates.PopShortList(K)
		}
		slog.Debug("List of candidates in order", "candidates.contacts", candidates.contacts)

		probed = 0
		for i := range len(candidates.contacts) {
			if slices.Contains(alreadyContacted, candidates.contacts[i]) {
				probed += 1
			}
		}

		if noNewClosest && (probed >= K || probed >= candidates.Len()) {
			break
		}
	}
	slog.Debug("Stopping the Lookup Loop", "!noNewClosest", !noNewClosest, "probed != k_const", probed != K, "valueFound", valueFound)
	slog.Debug("State of bucket", "k.RoutingTable.FindClosestContacts(k.me.ID, 10)", k.RoutingTable.FindClosestContacts(k.me.ID, 10))
	var contactWithAnswer Contact
	if answer != nil {
		for i := range len(candidates.contacts) {
			if !slices.Contains(candidatesWithAnswer, *candidates.contacts[i].ID) {
				k.SendStore(candidates.contacts[i], target, *answer)
			} else {
				contactWithAnswer = candidates.contacts[i]
				break
			}
		}
		if contactWithAnswer.ID == nil && len(alreadyContacted) > 0 {
			for i := range alreadyContacted {
				if alreadyContacted[i].ID != nil && slices.Contains(candidatesWithAnswer, *alreadyContacted[i].ID) {
					contactWithAnswer = alreadyContacted[i]
					break
				}
			}
		}
		return answer, &contactWithAnswer, nil, nil
	}

	return nil, nil, &candidates, nil
}

func (k *kademlia) Store(key *entities.KademliaID, data string) error {
	candidates, err := k.LookupContact(key)
	if candidates == nil {
		candidates = &ContactCandidates{}
	}
	if err != nil {
		return err
	}

	contactMe := k.me
	contactMe.CalcDistance(key)
	slog.Info("Adding myself", "[]Contact{contactMe}", []Contact{contactMe})
	candidates.Append([]Contact{contactMe})
	candidates.Sort()
	if candidates.Len() > K {
		candidates.PopShortList(K)
	}
	slog.Info("Candidates are :", "candidates.contacts", candidates.contacts)

	for i := range len(candidates.contacts) {
		if candidates.contacts[i].ID != nil && candidates.contacts[i].ID.Equals(k.me.ID) {
			_ = k.dataStore.Put(*key, data)
			continue
		}
		err := k.SendStore(candidates.contacts[i], key, data)
		if err != nil {
			return err
		}
	}

	return nil
}

func (k *kademlia) Ping(address entities.Address) (time.Duration, error) {
	connection, err := k.network.Dial(address)
	if err != nil {
		return 0, err
	}
	defer connection.Close()

	requestUUID := uuid.New().String()

	pingMessage := generated.Message{
		Contact: &generated.Contact{
			KademliaId: k.me.ID[:],
			Ip:         k.me.Address.IP,
			Port:       int32(k.me.Address.Port),
		},
		Payload: &generated.Message_Ping{
			Ping: &generated.Ping{
				RequestUuid: requestUUID,
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

	if requestUUID != pongMessage.RequestUuid {
		return 0, errors.New("invalid request uuid")
	}

	return elapsedTime, nil
}

func (k *kademlia) Put(data string) (*entities.KademliaID, error) {
	id := entities.NewKademliaIDFromString(data)
	err := k.Store(id, data)
	if err != nil {
		return nil, err
	}
	return id, nil
}

func (k *kademlia) GetBuckets() []*Bucket {
	return k.RoutingTable.getBuckets()
}

func (k *kademlia) GetStoredKeys() []entities.KademliaID {
	return k.dataStore.Keys()
}

func (k *kademlia) GetValue(key string) (*string, *string, error) {
	target, err := entities.NewKademliaID(key)
	if err != nil {
		return nil, nil, err
	}
	value, contact, _, err := k.LookupValue(target)
	if err != nil {
		return nil, nil, err

	}
	if value != nil {
		var contactID *string
		if contact != nil && contact.ID != nil {
			id := contact.ID.String()
			contactID = &id
		}
		return value, contactID, nil
	}
	return nil, nil, nil
}
