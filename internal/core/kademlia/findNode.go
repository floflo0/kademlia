package kademlia

import (
	"kademlia/internal/core/entities"
	"kademlia/internal/core/ports"
	"kademlia/proto/generated"
	"log/slog"

	"google.golang.org/protobuf/proto"
)

type RPCResponse struct {
	double []Double
}

type Double struct {
	address entities.Address
	id      *KademliaID
}

func SendFindNode(requester_id *KademliaID, net ports.Network, recipient Contact, target *KademliaID) (*RPCResponse, error) {
	slog.Debug("Sending find node", "recipient ID", recipient.ID)
	connection, errDial := net.Dial(recipient.Address)
	if errDial != nil {
		slog.Debug("Dial")
		return nil, errDial
	}

	findNodeMessage := generated.Message{
		KademliaId: requester_id[:],
		Payload: &generated.Message_FindNode{
			FindNode: &generated.FindNode{
				TargetId:    target[:],
				RequesterId: requester_id[:],
				RecipientId: recipient.ID[:],
			},
		},
	}

	out, errMarshal := proto.Marshal(&findNodeMessage)
	if errMarshal != nil {
		slog.Debug("Marshal")
		return nil, errMarshal
	}

	errSend := connection.Send(out)
	if errSend != nil {
		slog.Debug("Send")
		return nil, errSend
	}

	recv, errRcv := connection.Receive(500) //For the moment random value
	if errRcv != nil {
		return nil, errRcv
	}

	msgRecv := &generated.FindNodeResponse{}
	err := proto.Unmarshal(recv, msgRecv)
	if err != nil {
		return nil, err
	}

	dataRecv := msgRecv.GetTriples()
	var response RPCResponse
	for i := range len(dataRecv) {
		response.double = append(response.double, Double{
			entities.Address{
				IP:   dataRecv[i].GetAddress(),
				Port: int(dataRecv[i].GetPort()),
			},
			NewKademliaID(string(dataRecv[i].GetKademliaid())),
		})
	}

	return &response, nil
}
