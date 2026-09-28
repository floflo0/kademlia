package kademlia

import (
	"kademlia/internal/core/entities"
	"kademlia/internal/core/ports"
	"kademlia/proto/generated"
	"log/slog"

	"google.golang.org/protobuf/proto"
)

type RPCResponseNode struct {
	double []DoubleNode
}

type DoubleNode struct {
	address entities.Address
	id      *KademliaID
}

func SendFindNode(requester_id *KademliaID, net ports.Network, recipient Contact, target *KademliaID) (*RPCResponseNode, error) {
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

	recv, errRcv := connection.Receive(500) //For the moment random value for timeout
	if errRcv != nil {
		return nil, errRcv
	}

	msgRecv := &generated.Message{}
	err := proto.Unmarshal(recv, msgRecv)
	if err != nil {
		return nil, err
	}

	dataRecv := msgRecv.GetFindNodeReponse().GetTriples()
	var response RPCResponseNode
	for i := range len(dataRecv) {
		response.double = append(response.double, DoubleNode{
			entities.Address{
				IP:   dataRecv[i].GetAddress(),
				Port: int(dataRecv[i].GetPort()),
			},
			NewKademliaID(string(dataRecv[i].GetKademliaid())),
		})
	}

	return &response, nil
}
