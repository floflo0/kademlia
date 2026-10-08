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
	slog.Debug("Sending find value", "recipient ID", recipient.ID.String())
	connection, errDial := k.network.Dial(recipient.Address)
	if errDial != nil {
		slog.Debug("Dial")
		return nil, errDial
	}
	defer connection.Close()

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

	recv, errRcv := connection.Receive(1000)
	if errRcv != nil {
		return nil, errRcv
	}

	msgRecv := &generated.Message{}
	err := proto.Unmarshal(recv, msgRecv)
	if err != nil {
		return nil, err
	}

	var response RPCResponseValue

	var senderID entities.KademliaID
	if msgRecv.Contact != nil && len(msgRecv.Contact.KademliaId) >= entities.KademliaIDLength {
		copy(senderID[:], msgRecv.Contact.KademliaId)
		response.sender = &senderID
	}

	valueRecv := msgRecv.GetFindValueResponse().GetValue()
	if valueRecv != "" {
		response.double = nil
		response.value = &valueRecv
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

	return &response, nil
}
