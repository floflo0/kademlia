package kademlia

import (
	"kademlia/internal/core/entities"
	"kademlia/internal/core/ports"
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

func SendFindValue(
	requester_id *entities.KademliaID,
	net ports.Network,
	recipient Contact,
	target *entities.KademliaID,
) (*RPCResponseValue, error) {
	slog.Debug("Sending find value", "recipient ID", recipient.ID)
	connection, errDial := net.Dial(recipient.Address)
	if errDial != nil {
		slog.Debug("Dial")
		return nil, errDial
	}

	findValueMessage := generated.Message{
		KademliaId: requester_id[:],
		Payload: &generated.Message_FindValue{
			FindValue: &generated.FindValue{
				TargetId:    target[:],
				RequesterId: requester_id[:],
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
		response.sender = (*entities.KademliaID)(msgRecv.GetKademliaId())
		return &response, nil
	}

	var doubleForResponse []DoubleValue
	value := ""
	dataRecv := msgRecv.GetFindValueResponse().GetTriples()
	for i := range len(dataRecv) {
		id, _ := entities.NewKademliaID(string(dataRecv[i].GetKademliaid()))
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
	response.sender = (*entities.KademliaID)(msgRecv.GetKademliaId())

	return &response, nil
}
