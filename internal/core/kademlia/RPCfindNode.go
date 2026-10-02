package kademlia

import (
	"kademlia/internal/core/entities"
	"kademlia/proto/generated"
	"log/slog"

	"google.golang.org/protobuf/proto"
)

type RPCResponseNode struct {
	double []DoubleNode
}

type DoubleNode struct {
	address entities.Address
	id      *entities.KademliaID
}

func (k *kademlia) SendFindNode(
	recipient Contact,
	target *entities.KademliaID,
) (*RPCResponseNode, error) {
	slog.Debug(
		"Sending FIND_NODE RPC",
		"key",
		target.String(),
		"to",
		recipient.Address,
	)
	connection, errDial := k.network.Dial(recipient.Address)
	if errDial != nil {
		slog.Debug("Dial")
		return nil, errDial
	}

	findNodeMessage := generated.Message{
		Contact: &generated.Contact{
			KademliaId: k.me.ID[:],
			Ip:         k.me.Address.IP,
			Port:       int32(k.me.Address.Port),
		},
		Payload: &generated.Message_FindNode{
			FindNode: &generated.FindNode{
				Key: target[:],
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
		id, err := entities.NewKademliaID(string(dataRecv[i].GetKademliaid()))
		if err != nil {
			return nil, err
		}
		response.double = append(response.double, DoubleNode{
			entities.Address{
				IP:   dataRecv[i].GetAddress(),
				Port: int(dataRecv[i].GetPort()),
			},
			id,
		})
	}

	return &response, nil
}
