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

func SendFindNode(net ports.Network, address entities.Address, target *KademliaID) (*RPCResponse, error) {
	connection, errDial := net.Dial(address)
	if errDial != nil {
		slog.Debug("Dial")
		return nil, errDial
	}

	var message generated.FindNode
	message.Data = target[:]

	out, errMarshal := proto.Marshal(&message)
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
