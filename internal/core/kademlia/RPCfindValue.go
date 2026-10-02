package kademlia

import (
	"kademlia/internal/core/entities"
	"kademlia/proto/generated"
	"log/slog"

	"google.golang.org/protobuf/proto"
)

type RPCResponseValue struct {
	double *[]DoubleValue
	value  *string
	sender *entities.KademliaID
}

type DoubleValue struct {
	address entities.Address
	id      *entities.KademliaID
}

func (k *kademlia) SendFindValue(
	recipient Contact,
	target *entities.KademliaID,
) (*RPCResponseValue, error) {
	slog.Debug("Sending find value", "recipient ID", recipient.ID)
	connection, errDial := k.network.Dial(recipient.Address)
	if errDial != nil {
		slog.Debug("Dial")
		return nil, errDial
	}

	findValueMessage := generated.Message{
		Contact: &generated.Contact{
			KademliaId: k.me.ID[:],
			Ip:         k.me.Address.IP,
			Port:       int32(k.me.Address.Port),
		},
		Payload: &generated.Message_FindValue{
			FindValue: &generated.FindValue{
				TargetId:    target[:],
				RequesterId: k.me.ID[:],
				RecipientId: recipient.ID[:],
			},
		},
	}

	out, errMarshal := proto.Marshal(&findValueMessage)
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

	var response RPCResponseValue

	valueRecv := msgRecv.GetFindValueResponse().GetValue()
	if valueRecv != "" {
		response.double = nil
		response.value = &valueRecv
		response.sender = (*entities.KademliaID)(msgRecv.Contact.KademliaId)
		return &response, nil
	}

	var doubleForResponse []DoubleValue
	value := ""
	dataRecv := msgRecv.GetFindValueResponse().GetTriples()
	for i := range len(dataRecv) {
		id, err := entities.NewKademliaID(string(dataRecv[i].GetKademliaid()))
		if err != nil {
			return nil, err
		}
		doubleForResponse = append(doubleForResponse, DoubleValue{
			address: entities.Address{
				IP:   dataRecv[i].GetAddress(),
				Port: int(dataRecv[i].GetPort()),
			},
			id: id,
		})
	}
	response.double = &doubleForResponse
	response.value = &value
	response.sender = (*entities.KademliaID)(msgRecv.Contact.KademliaId)

	return &response, nil
}
