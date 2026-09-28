package kademlia

import (
	"errors"
	"kademlia/internal/core/ports"
	"kademlia/proto/generated"
	"log/slog"

	"google.golang.org/protobuf/proto"
)

var ErrTooMuchData = errors.New("Too much data, do not exceed 1024 bytes")

func SendStore(requester_id *KademliaID, net ports.Network, recipient Contact, key string, data string) error {
	slog.Debug("Sending Store", "recipient ID", recipient.ID)
	connection, errDial := net.Dial(recipient.Address)
	if errDial != nil {
		slog.Debug("Dial")
		return errDial
	}

	if len([]byte(data)) > 1024 {
		return ErrTooMuchData
	}

	storeMessage := generated.Message{
		KademliaId: requester_id[:],
		Payload: &generated.Message_Store{
			Store: &generated.Store{
				Key:         []byte(key),
				Data:        []byte(data),
				DataLen:     int32(len(data)),
				RequesterId: requester_id[:],
				RecipientId: recipient.ID[:],
			},
		},
	}

	out, errMarshal := proto.Marshal(&storeMessage)
	if errMarshal != nil {
		slog.Debug("Marshal")
		return errMarshal
	}

	errSend := connection.Send(out)
	if errSend != nil {
		slog.Debug("Send")
		return errSend
	}

	return nil
}
